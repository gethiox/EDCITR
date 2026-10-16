package zaputils

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	_ "github.com/gethiox/EDCITR/internal/console" // crucial anonymous import to ensure correct os.Stdout file descriptor for this package
	"github.com/gethiox/EDCITR/internal/utils"
	"github.com/gethiox/rotational-velocidensity-buffer"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

func init() {
	stdout := zapcore.AddSync(os.Stdout)

	bufferedOut := zapcore.AddSync(globalLogWriter)

	cfg, cfgRaw := getEncoderConfigs()
	cfg.EncodeCaller = nil
	cfgRaw.EncodeCaller = nil

	baseEncoder := zapcore.NewConsoleEncoder(cfg)
	baseEncoderRaw := zapcore.NewConsoleEncoder(cfgRaw)

	coloredEncoder := &ColorizedEncoder{Encoder: baseEncoder}
	coloredEncoderRaw := &ColorizedEncoder{Encoder: baseEncoderRaw}

	core := zapcore.NewTee(
		zapcore.NewCore(coloredEncoder, stdout, atomLevel),
		zapcore.NewCore(coloredEncoder, bufferedOut, atomLevel),
	)

	coreRaw := zapcore.NewTee(
		zapcore.NewCore(coloredEncoderRaw, stdout, atomLevel),
		zapcore.NewCore(coloredEncoderRaw, bufferedOut, atomLevel),
	)

	Logger = zap.New(core, zap.AddCaller())
	LoggerWithoutCaller = zap.New(core)
	LoggerRaw = zap.New(coreRaw)
}

var (
	Logger              *zap.Logger
	LoggerWithoutCaller *zap.Logger
	LoggerRaw           *zap.Logger

	globalLogWriter = &HotSwapWriter{}
	atomLevel       = zap.NewAtomicLevelAt(zap.InfoLevel)
)

func SetLevel(l zapcore.Level) {
	atomLevel.SetLevel(l)
}

func ConnectLogBuffer(rvb *rvb.RVBuffer[string]) {
	globalLogWriter.AttachTarget(rvb)
}

type HotSwapWriter struct {
	mtx       sync.RWMutex
	earlyLogs []string
	target    *rvb.RVBuffer[string]
}

func (w *HotSwapWriter) Write(p []byte) (int, error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	msg := string(p)

	if w.target != nil {
		w.target.Push(msg)
	} else {
		w.earlyLogs = append(w.earlyLogs, msg)
	}

	return len(p), nil
}

func (w *HotSwapWriter) Sync() error {
	return nil
}

func (w *HotSwapWriter) AttachTarget(buf *rvb.RVBuffer[string]) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	if w.target != nil {
		return // already attached
	}

	// replay history
	if len(w.earlyLogs) > 0 {
		buf.PushMany(w.earlyLogs)
		w.earlyLogs = nil
	}

	w.target = buf
}

func CustomLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var color string
	switch l {
	case zapcore.DebugLevel:
		color = colorGray
	case zapcore.InfoLevel:
		color = colorGreen
	case zapcore.WarnLevel:
		color = colorYellow
	case zapcore.ErrorLevel:
		color = colorRed
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		color = colorBrightRed
	default:
		color = colorWhite
	}

	enc.AppendString(fmt.Sprintf("%s%5s%s", color, l.CapitalString(), colorReset))
}

func CustomNameEncoder(name string, enc zapcore.PrimitiveArrayEncoder) {
	// 1. Calculate a simple hash of the string to pick a color index
	hash := 0
	for i := 0; i < len(name); i++ {
		hash = 31*hash + int(name[i])
	}
	if hash < 0 {
		hash = -hash
	}

	// 2. Define a pool of "nice" colors (Codes 31-36 and 91-96)
	// We skip 30/90 (Black/Gray) and 37/97 (White) to keep names distinct.
	colors := []string{
		colorRed,
		colorGreen,
		colorYellow,
		colorBlue,
		colorMagenta,
		colorCyan,
		colorBrightRed,
		colorBrightGreen,
		colorBrightYellow,
		colorBrightBlue,
		colorBrightMagenta,
		colorBrightCyan,
	}

	// 3. Pick the color
	colorCode := colors[hash%len(colors)]

	// 4. Format: [COLOR]name[RESET]
	enc.AppendString(fmt.Sprintf("%s[%s]%s", colorCode, name, colorReset))
}

func getEncoderConfigs() (cfg, cfgRaw zapcore.EncoderConfig) {
	cfg = zap.NewDevelopmentEncoderConfig()
	cfg.ConsoleSeparator = " "
	cfg.EncodeTime = ColoredTimeEncoder
	cfg.EncodeLevel = CustomLevelEncoder
	cfg.EncodeName = CustomNameEncoder

	cfgRaw = zap.NewDevelopmentEncoderConfig()
	cfgRaw.LevelKey = ""
	cfgRaw.TimeKey = ""
	cfgRaw.CallerKey = ""
	cfgRaw.StacktraceKey = ""
	cfgRaw.FunctionKey = ""
	cfgRaw.NameKey = ""

	return cfg, cfgRaw
}

type BufferWriter struct {
	*rvb.RVBuffer[string]
}

func (w BufferWriter) Write(p []byte) (n int, err error) {
	w.Push(string(p))
	return len(p), nil
}

func (w BufferWriter) Sync() error {
	return nil
}

const (
	colorReset         = "\x1b[0m"
	colorRed           = "\x1b[31m"
	colorGreen         = "\x1b[32m"
	colorYellow        = "\x1b[33m"
	colorBlue          = "\x1b[34m"
	colorPurple        = "\x1b[35m"
	colorMagenta       = colorPurple
	colorCyan          = "\x1b[36m"
	colorGray          = colorBrightBlack
	colorWhite         = "\x1b[37m"
	colorBrightBlack   = "\x1b[90m"
	colorBrightRed     = "\x1b[91m"
	colorBrightGreen   = "\x1b[92m"
	colorBrightYellow  = "\x1b[93m"
	colorBrightBlue    = "\x1b[94m"
	colorBrightMagenta = "\x1b[95m"
	colorBrightCyan    = "\x1b[96m"
)

func ColoredTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	atenc, ok := enc.(appendTimeEncoder)
	if ok {
		atenc.AppendTimeLayout(t, "2006-01-02 15:04:05.000")
		return
	}

	formatted := fmt.Sprintf("%s%s %s%s%s%s%s",
		colorCyan, t.Format("2006-01-02"),
		colorGreen, t.Format("15:04:05"),
		colorGray, t.Format(".000"),
		colorReset,
	)

	enc.AppendString(formatted)
}

type ColorizedEncoder struct {
	zapcore.Encoder
	context []zapcore.Field
}

func (e *ColorizedEncoder) Clone() zapcore.Encoder {
	newCtx := make([]zapcore.Field, len(e.context))
	copy(newCtx, e.context)

	return &ColorizedEncoder{
		Encoder: e.Encoder.Clone(),
		context: newCtx,
	}
}

func (e *ColorizedEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf, err := e.Encoder.EncodeEntry(entry, nil)
	if err != nil {
		return nil, err
	}
	buf.TrimNewline()

	processFields := func(fields []zapcore.Field) {
		for _, f := range fields {
			buf.AppendString(" ")
			e.encodeColoredField(buf, f)
		}
	}

	processFields(e.context) // .With() fields
	processFields(fields)    // .Info() fields

	if entry.Caller.Defined {
		buf.AppendString(" ")
		buf.AppendString(colorGray)
		buf.AppendString("<")
		buf.AppendString(entry.Caller.TrimmedPath())
		buf.AppendString(">")
		buf.AppendString(colorReset)
	}

	buf.AppendString("\n")
	return buf, nil
}

func (e *ColorizedEncoder) encodeColoredField(buf *buffer.Buffer, f zapcore.Field) {
	buf.AppendString(colorPurple)
	buf.AppendString(f.Key)
	buf.AppendString(colorReset)
	buf.AppendString("=")

	switch f.Type {
	case zapcore.StringType:
		buf.AppendString(colorGreen)
		buf.AppendString(fmt.Sprintf("'%s'", f.String))
	case zapcore.Int64Type, zapcore.Int32Type, zapcore.Uint64Type, zapcore.Uint32Type:
		buf.AppendString(colorBlue)
		buf.AppendInt(f.Integer)
	case zapcore.Float32Type:
		buf.AppendString(colorBlue)
		val := math.Float32frombits(uint32(f.Integer))
		buf.AppendString(fmt.Sprintf("%.2f", val))
	case zapcore.Float64Type:
		buf.AppendString(colorBlue)
		val := math.Float64frombits(uint64(f.Integer))
		buf.AppendString(fmt.Sprintf("%.2f", val))
	case zapcore.DurationType:
		buf.AppendString(colorCyan)
		buf.AppendString(utils.FormatDurationPrecise(time.Duration(f.Integer)))
	case zapcore.TimeType:
		buf.AppendString(colorCyan)
		switch t := f.Interface.(type) {
		case time.Time:
			buf.AppendString(t.Format(time.RFC3339))
		default:
			tt := time.Unix(0, f.Integer)
			buf.AppendString(tt.Format(time.RFC3339))
		}
	case zapcore.BoolType:
		buf.AppendString(colorBlue)
		if f.Integer == 1 {
			buf.AppendString("true")
		} else {
			buf.AppendString("false")
		}
	case zapcore.ErrorType:
		buf.AppendString(colorRed)
		if err, ok := f.Interface.(error); ok && err != nil {
			buf.AppendString(fmt.Sprintf("'%s'", err.Error()))
		} else {
			buf.AppendString("nil")
		}
	case zapcore.ArrayMarshalerType:
		buf.AppendString(colorGreen)

		if f.Interface == nil {
			buf.AppendString("[]")
		} else if arr, ok := f.Interface.(zapcore.ArrayMarshaler); ok {
			enc := zapcore.NewMapObjectEncoder()
			_ = enc.AddArray("tmp", arr)

			extractedSlice, exists := enc.Fields["tmp"]
			if exists {
				buf.AppendString(fmt.Sprintf("%v", extractedSlice))
			} else {
				buf.AppendString("[]")
			}
		} else {
			buf.AppendString(fmt.Sprintf("%v", f.Interface))
		}
	default:
		buf.AppendString(colorGray)
		if f.Interface == nil {
			buf.AppendString(fmt.Sprintf("<unk type:%v val:%#v>", f.Type, f.Integer))
		} else {
			buf.AppendString(fmt.Sprintf("%#v", f.Interface))
		}
	}

	buf.AppendString(colorReset)
}
