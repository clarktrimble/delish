package value

import (
	"strconv"
)

type Value struct {
	Data   []byte
	Quoted bool
}

/*
func New(value []byte) Value {

	return Value{
		Data: value,
	}
}
*/

func NewFromString(str string) (val Value) {

	// Todo: use encoder with reusable buf, but where buf??
	//       or can I use strconv AppendQuote?
	//       how to properly concur?

	//data, _ := json.Marshal(value)
	//return Value{
	//Data: data,
	//}
	val = Value{
		Data: make([]byte, 0, len(str)+2),
		//Data: []byte{},
	}
	//b := make([]byte, 0, 1024)
	val.Data = strconv.AppendQuote(val.Data, str)
	val.Quoted = true

	return
}

func (vl Value) String() string {
	return string(vl.Data)
}

func (vl Value) MarshalJSON() ([]byte, error) {

	//fmt.Printf(">>> Data:%s\n", vl.Data)

	return vl.Data, nil
}

//var ValuePool = sync.Pool{
//New: func() any {
//return Value{Data: make([]byte, 0, 999)}
//},
//}

//func (vl Value) Write(data []byte) (int, error) {
//*vl.Data = append(*vl.Data, data...)
//return len(*vl.Data), nil
//}
