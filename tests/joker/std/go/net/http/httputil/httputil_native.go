// Auto-generated by gostd2joker at (omitted for testing), do not edit!!

package httputil

import (
)

// func dumpRequest(req ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/httputil/dump.go:191:22), body bool) Object {
// 	_res1, _res2 := _httputil.DumpRequest(req, body)
// 	_res := EmptyVector
// 	_vec1 := EmptyVector
// 	for _, _elem1 := range _res1 {
// 		_vec1 = _vec1.Conjoin(MakeInt(int(_elem1)))
// 	}
// 	_res = _res.Conjoin(_vec1)
// 	_res = _res.Conjoin(func () Object { if (_res2) == nil { return NIL } else { return MakeError(_res2) } }())
// 	return _res
// }

// func dumpRequestOut(req ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/httputil/dump.go:66:25), body bool) Object {
// 	_res1, _res2 := _httputil.DumpRequestOut(req, body)
// 	_res := EmptyVector
// 	_vec1 := EmptyVector
// 	for _, _elem1 := range _res1 {
// 		_vec1 = _vec1.Conjoin(MakeInt(int(_elem1)))
// 	}
// 	_res = _res.Conjoin(_vec1)
// 	_res = _res.Conjoin(func () Object { if (_res2) == nil { return NIL } else { return MakeError(_res2) } }())
// 	return _res
// }

// func dumpResponse(resp ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/httputil/dump.go:281:24), body bool) Object {
// 	_res1, _res2 := _httputil.DumpResponse(resp, body)
// 	_res := EmptyVector
// 	_vec1 := EmptyVector
// 	for _, _elem1 := range _res1 {
// 		_vec1 = _vec1.Conjoin(MakeInt(int(_elem1)))
// 	}
// 	_res = _res.Conjoin(_vec1)
// 	_res = _res.Conjoin(func () Object { if (_res2) == nil { return NIL } else { return MakeError(_res2) } }())
// 	return _res
// }

// func newChunkedReader(r ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/http/httputil/httputil.go:20:25)) Object {
// 	return _httputil.NewChunkedReader(r)
// }

// func newChunkedWriter(w ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/http/httputil/httputil.go:35:25)) Object {
// 	return _httputil.NewChunkedWriter(w)
// }

// func newClientConn(c ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/http/httputil/persist.go:248:22), r ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/httputil/persist.go:248:34)) Object {
// 	return _httputil.NewClientConn(c, r)
// 	ABEND124(no public information returned)
// }

// func newProxyClientConn(c ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/http/httputil/persist.go:265:27), r ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/httputil/persist.go:265:39)) Object {
// 	return _httputil.NewProxyClientConn(c, r)
// 	ABEND124(no public information returned)
// }

// func newServerConn(c ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/http/httputil/persist.go:54:22), r ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/httputil/persist.go:54:34)) Object {
// 	return _httputil.NewServerConn(c, r)
// 	ABEND124(no public information returned)
// }

// func newSingleHostReverseProxy(target ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/httputil/reverseproxy.go:103:39)) Object {
// 	_res := _httputil.NewSingleHostReverseProxy(target)
// 	var _obj_map1 Object
// 	if _res != nil {
// 		_map1 := EmptyArrayMap()
// 		_map1.Add(MakeKeyword("Director"), (*_res).Director)
// 		_map1.Add(MakeKeyword("Transport"), (*_res).Transport)
// 		_map1.Add(MakeKeyword("FlushInterval"), (*_res).FlushInterval)
// 		_map1.Add(MakeKeyword("ErrorLog"), (*(*_res).ErrorLog))
// 		_map1.Add(MakeKeyword("BufferPool"), (*_res).BufferPool)
// 		_map1.Add(MakeKeyword("ModifyResponse"), (*_res).ModifyResponse)
// 		_map1.Add(MakeKeyword("ErrorHandler"), (*_res).ErrorHandler)
// 		_obj_map1 = Object(_map1)
// 	} else {
// 		_obj_map1 = NIL
// 	}
// 	return _obj_map1
// }
