package inject

var (
	TH32CS_SNAPPROCESS  uint32 = 0x00000002
	TH32CS_SNAPTHREAD   uint32 = 0x00000004
	THREAD_ALL_ACCESS   uint32 = 0xffff
	ERROR_NO_MORE_FILES string = "There are no more files."
	SUCCESS             string = "The operation completed successfully."
	CONTEXT_FULL        uint32 = 0x400003
)

type HOOKPROC func(int, uintptr, uintptr) uintptr

type HANDLER func() uintptr

type baseRelocEntry uint16

func (b baseRelocEntry) Type() uint16 {
	return uint16(uint16(b) >> 12)
}

func (b baseRelocEntry) Offset() uint32 {
	return uint32(uint16(b) & 0x0FFF)
}

// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-context
type XMM_SAVE_AREA32 struct {
	ControlWord    uint16
	StatusWord     uint16
	TagWord        byte
	Reserved1      byte
	ErrorOpcode    uint16
	ErrorOffset    uint32
	ErrorSelector  uint16
	Reserved2      uint16
	DataOffset     uint32
	DataSelector   uint16
	Reserved3      uint16
	MxCsr          uint32
	MxCsr_Mask     uint32
	FloatRegisters [8]M128A
	XmmRegisters   [256]byte
	Reserved4      [96]byte
}

// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-context
type M128A struct {
	Low  uint64
	High int64
}

// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-context
type CONTEXT struct {
	P1Home uint64
	P2Home uint64
	P3Home uint64
	P4Home uint64
	P5Home uint64
	P6Home uint64

	ContextFlags uint32
	MxCsr        uint32

	SegCs  uint16
	SegDs  uint16
	SegEs  uint16
	SegFs  uint16
	SegGs  uint16
	SegSs  uint16
	EFlags uint32

	Dr0 uint64
	Dr1 uint64
	Dr2 uint64
	Dr3 uint64
	Dr6 uint64
	Dr7 uint64

	Rax uint64
	Rcx uint64
	Rdx uint64
	Rbx uint64
	Rsp uint64
	Rbp uint64
	Rsi uint64
	Rdi uint64
	R8  uint64
	R9  uint64
	R10 uint64
	R11 uint64
	R12 uint64
	R13 uint64
	R14 uint64
	R15 uint64

	Rip uint64

	FltSave XMM_SAVE_AREA32

	VectorRegister [26]M128A
	VectorControl  uint64

	DebugControl         uint64
	LastBranchToRip      uint64
	LastBranchFromRip    uint64
	LastExceptionToRip   uint64
	LastExceptionFromRip uint64
}
