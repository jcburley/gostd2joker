// Auto-generated by gostd2joker at (omitted for testing), do not edit!!

package textproto

import (
	"net/textproto"
	. "github.com/candid82/joker/core"
)

func dial(network string, addr string) Object {
	res1, res2 := textproto.Dial(network, addr)
	res := EmptyVector
	res = res.Conjoin(NIL)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

// func newConn(conn ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/textproto/textproto.go:66:19)) Object {
// 	return textproto.NewConn(conn)
// 	ABEND124(no public information returned)
// }

// func newReader(r ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/textproto/reader.go:29:18)) Object {
// 	res := textproto.NewReader(r)
// 	map1 := EmptyArrayMap()
// 	map1.Add(MakeKeyword("R"), (*(*res).R))
// 	return map1
// }

// func newWriter(w ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/textproto/writer.go:21:18)) Object {
// 	res := textproto.NewWriter(w)
// 	map1 := EmptyArrayMap()
// 	map1.Add(MakeKeyword("W"), (*(*res).W))
// 	return map1
// }

// func trimBytes(b ABEND882(unrecognized Expr type *ast.ArrayType at: tests/big/src/net/textproto/textproto.go:137:18)) Object {
// 	res := textproto.TrimBytes(b)
// 	vec1 := EmptyVector
// 	for _, elem1 := range res {
// 		vec1 = vec1.Conjoin(MakeInt(int(elem1)))
// 	}
// 	return vec1
// }
