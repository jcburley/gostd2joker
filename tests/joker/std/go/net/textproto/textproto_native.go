// Auto-generated by gostd2joker at (omitted for testing), do not edit!!

package textproto

import (
	_textproto "net/textproto"
	. "github.com/candid82/joker/core"
)

func dial(network string, addr string) Object {
	_, _res2 := _textproto.Dial(network, addr)
	_res := EmptyVector
	_res = _res.Conjoin(NIL)
	_res = _res.Conjoin(func () Object { if (_res2) == nil { return NIL } else { return MakeError(_res2) } }())
	return _res
}

// func newConn(conn ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/textproto/textproto.go:66:19)) Object {
// 	return _textproto.NewConn(conn)
// 	ABEND124(no public information returned)
// }

// func newReader(r ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/textproto/reader.go:29:18)) Object {
// 	_res := _textproto.NewReader(r)
// 	_map1 := EmptyArrayMap()
// 	_map1.Add(MakeKeyword("R"), (*(*_res).R))
// 	return _map1
// }

// func newWriter(w ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/textproto/writer.go:21:18)) Object {
// 	_res := _textproto.NewWriter(w)
// 	_map1 := EmptyArrayMap()
// 	_map1.Add(MakeKeyword("W"), (*(*_res).W))
// 	return _map1
// }

// func trimBytes(b ABEND882(unrecognized Expr type *ast.ArrayType at: tests/big/src/net/textproto/textproto.go:137:18)) Object {
// 	_res := _textproto.TrimBytes(b)
// 	_vec1 := EmptyVector
// 	for _, _elem1 := range _res {
// 		_vec1 = _vec1.Conjoin(MakeInt(int(_elem1)))
// 	}
// 	return _vec1
// }
