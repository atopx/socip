package socdb

const noDataMapKey dataMapKey = ""
const noOffset = ^uintptr(0)

const (
	typeNumExtended typeNum = iota
	typeNumPointer
	typeNumString
	typeNumFloat64
	typeNumBytes
	typeNumUint16
	typeNumUint32
	typeNumMap
	typeNumInt32
	typeNumUint64
	typeNumUint128
	typeNumSlice
	// We don't use the next two. They are placeholders. See the spec
	// for more details.
	typeNumContainer // nolint: deadcode, varcheck
	typeNumMarker    // nolint: deadcode, varcheck
	typeNumBool
	typeNumFloat32
)

const (
	recordTypeEmpty recordType = iota
	recordTypeData
	recordTypeNode
	recordTypeAlias
	recordTypeFixedNode
	recordTypeReserved
)
