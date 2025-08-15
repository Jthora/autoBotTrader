package ephem

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

// Bitmask for fields present in GTAB data file.
const (
    FieldTideBPS      uint32 = 0x01
    FieldTideRawF32   uint32 = 0x02
    FieldMoonRkmF32   uint32 = 0x04
    FieldSunRkmF32    uint32 = 0x08
    FieldMoonRinv3F32 uint32 = 0x10
    FieldSunRinv3F32  uint32 = 0x20
)

// GTAB is a minimal reader for the GTAB v1 binary layout described in docs/EPHEMERIS_DATA_FORMAT.md.
// It supports O(1) indexed access and direct ReadAt without loading entire file into memory.
type GTAB struct {
    f          *os.File
    path       string
    epoch      time.Time
    dt         time.Duration
    n          uint32
    fieldsMask uint32
    headerSize int64
    recordSize int64
    offTideBPS int64 // -1 if absent
}

// Open opens a GTAB file and parses the header.
func Open(path string) (*GTAB, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    // Layout: magic[5] "GTAB1", version u16, epoch i64, dt_ns i64, n u32, fields_mask u32, reserved[16]
    hdr := make([]byte, 5+2+8+8+4+4+16)
    if _, err := io.ReadFull(f, hdr); err != nil {
        f.Close()
        return nil, fmt.Errorf("read header: %w", err)
    }
    if string(hdr[:5]) != "GTAB1" {
        f.Close()
        return nil, fmt.Errorf("%s: invalid GTAB magic", path)
    }
    ver := binary.LittleEndian.Uint16(hdr[5:7])
    if ver != 1 {
        f.Close()
        return nil, fmt.Errorf("%s: unsupported GTAB version: %d", path, ver)
    }
    epochSec := int64(binary.LittleEndian.Uint64(hdr[7:15]))
    dtNS := int64(binary.LittleEndian.Uint64(hdr[15:23]))
    n := binary.LittleEndian.Uint32(hdr[23:27])
    fields := binary.LittleEndian.Uint32(hdr[27:31])

    if dtNS <= 0 {
        f.Close()
        return nil, fmt.Errorf("%s: invalid dt_ns: %d", path, dtNS)
    }
    if n == 0 {
        f.Close()
        return nil, fmt.Errorf("%s: empty table (n=0)", path)
    }

    // Determine record layout in fixed order
    var recSize int64 = 0
    offTideBPS := int64(-1)
    if fields&FieldTideBPS != 0 {
        offTideBPS = recSize
        recSize += 2
    }
    if fields&FieldTideRawF32 != 0 {
        recSize += 4
    }
    if fields&FieldMoonRkmF32 != 0 {
        recSize += 4
    }
    if fields&FieldSunRkmF32 != 0 {
        recSize += 4
    }
    if fields&FieldMoonRinv3F32 != 0 {
        recSize += 4
    }
    if fields&FieldSunRinv3F32 != 0 {
        recSize += 4
    }
    if recSize == 0 {
        f.Close()
        return nil, fmt.Errorf("%s: empty record layout (fields_mask=0)", path)
    }
    // Validate file size matches header + n*recordSize
    st, err := f.Stat()
    if err != nil {
        f.Close()
        return nil, fmt.Errorf("stat file: %w", err)
    }
    expected := int64(len(hdr)) + int64(n)*recSize
    if st.Size() < expected {
        f.Close()
        return nil, fmt.Errorf("%s: truncated GTAB file: have %d bytes, expected %d", path, st.Size(), expected)
    }

    g := &GTAB{
        f:          f,
        path:       path,
        epoch:      time.Unix(epochSec, 0).UTC(),
        dt:         time.Duration(dtNS),
        n:          n,
        fieldsMask: fields,
        headerSize: int64(len(hdr)),
        recordSize: recSize,
        offTideBPS: offTideBPS,
    }
    return g, nil
}

// Close closes the underlying file.
func (g *GTAB) Close() error {
    if g.f != nil {
        return g.f.Close()
    }
    return nil
}

// Coverage returns the inclusive time window covered by the table.
func (g *GTAB) Coverage() (start, end time.Time) {
    if g.n == 0 {
        return g.epoch, g.epoch
    }
    end = g.epoch.Add(time.Duration(int64(g.n-1)) * g.dt)
    return g.epoch, end
}

// IndexFor returns the index for timestamp t and fractional position within the interval.
// If out of range, it returns -1 and false.
func (g *GTAB) IndexFor(t time.Time) (int64, float64, bool) {
    if g.n == 0 {
        return -1, 0, false
    }
    dt := t.UTC().Sub(g.epoch)
    pos := float64(dt) / float64(g.dt)
    if pos < 0 || pos > float64(g.n-1) {
        return -1, 0, false
    }
    i := int64(pos)
    frac := pos - float64(i)
    return i, frac, true
}

// readTideBPS reads the tide_bps value at index i. ok=false if field absent or read fails.
func (g *GTAB) readTideBPS(i int64) (uint16, bool) {
    if g.offTideBPS < 0 {
        return 0, false
    }
    off := g.headerSize + i*g.recordSize + g.offTideBPS
    buf := make([]byte, 2)
    if _, err := g.f.ReadAt(buf, off); err != nil {
        return 0, false
    }
    return binary.LittleEndian.Uint16(buf), true
}

// LookupTideBPS returns the linearly interpolated tide_bps at time t.
// ok=false when t is out of range or field missing.
func (g *GTAB) LookupTideBPS(t time.Time) (uint16, bool) {
    i, frac, ok := g.IndexFor(t)
    if !ok {
        return 0, false
    }
    if frac == 0 {
        v, ok := g.readTideBPS(i)
        return v, ok
    }
    v0, ok0 := g.readTideBPS(i)
    v1, ok1 := g.readTideBPS(i+1)
    if !ok0 || !ok1 {
        return 0, false
    }
    a := float64(v0)
    b := float64(v1)
    vf := a + (b-a)*frac
    if vf < 0 {
        vf = 0
    }
    if vf > 65535 {
        vf = 65535
    }
    return uint16(vf + 0.5), true
}
