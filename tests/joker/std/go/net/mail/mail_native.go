// Auto-generated by gostd2joker at (omitted for testing), do not edit!!

package mail

import (
	_mail "net/mail"
	. "github.com/candid82/joker/core"
)

func parseAddress(address string) Object {
	_res1, _res2 := _mail.ParseAddress(address)
	_res := EmptyVector
	_map1 := EmptyArrayMap()
	_map1.Add(MakeKeyword("Name"), MakeString((*_res1).Name))
	_map1.Add(MakeKeyword("Address"), MakeString((*_res1).Address))
	_res = _res.Conjoin(_map1)
	_res = _res.Conjoin(func () Object { if (_res2) == nil { return NIL } else { return MakeError(_res2) } }())
	return _res
}

func parseAddressList(list string) Object {
	_res1, _res2 := _mail.ParseAddressList(list)
	_res := EmptyVector
	_vec1 := EmptyVector
	for _, _elem1 := range _res1 {
		_map2 := EmptyArrayMap()
		_map2.Add(MakeKeyword("Name"), MakeString((*_elem1).Name))
		_map2.Add(MakeKeyword("Address"), MakeString((*_elem1).Address))
		_vec1 = _vec1.Conjoin(_map2)
	}
	_res = _res.Conjoin(_vec1)
	_res = _res.Conjoin(func () Object { if (_res2) == nil { return NIL } else { return MakeError(_res2) } }())
	return _res
}

// func parseDate(date string) Object {
// 	_res1, _res2 := _mail.ParseDate(date)
// 	_res := EmptyVector
// 	_res = _res.Conjoin(_res1)
// 	_res = _res.Conjoin(func () Object { if (_res2) == nil { return NIL } else { return MakeError(_res2) } }())
// 	return _res
// }

// func readMessage(r ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/mail/message.go:52:20)) Object {
// 	msg, err := _mail.ReadMessage(r)
// 	_res := EmptyVector
// 	_map1 := EmptyArrayMap()
// 	_map1.Add(MakeKeyword("Header"), (*msg).Header)
// 	_map1.Add(MakeKeyword("Body"), (*msg).Body)
// 	_res = _res.Conjoin(_map1)
// 	_res = _res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeError(err) } }())
// 	return _res
// }
