// Auto-generated by gostd2joker at (omitted for testing), do not edit!!

package url

import (
	"net/url"
	. "github.com/candid82/joker/core"
)

func parse(rawurl string) Object {
	res1, res2 := url.Parse(rawurl)
	res := EmptyVector
	map1 := EmptyArrayMap()
	map1.Add(MakeKeyword("Scheme"), MakeString((*res1).Scheme))
	map1.Add(MakeKeyword("Opaque"), MakeString((*res1).Opaque))
	map1.Add(MakeKeyword("User"), NIL)
	map1.Add(MakeKeyword("Host"), MakeString((*res1).Host))
	map1.Add(MakeKeyword("Path"), MakeString((*res1).Path))
	map1.Add(MakeKeyword("RawPath"), MakeString((*res1).RawPath))
	map1.Add(MakeKeyword("ForceQuery"), MakeBool((*res1).ForceQuery))
	map1.Add(MakeKeyword("RawQuery"), MakeString((*res1).RawQuery))
	map1.Add(MakeKeyword("Fragment"), MakeString((*res1).Fragment))
	res = res.Conjoin(map1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

// func parseQuery(query string) Object {
// 	res1, res2 := url.ParseQuery(query)
// 	res := EmptyVector
// 	res = res.Conjoin(res1)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

func parseRequestURI(rawurl string) Object {
	res1, res2 := url.ParseRequestURI(rawurl)
	res := EmptyVector
	map1 := EmptyArrayMap()
	map1.Add(MakeKeyword("Scheme"), MakeString((*res1).Scheme))
	map1.Add(MakeKeyword("Opaque"), MakeString((*res1).Opaque))
	map1.Add(MakeKeyword("User"), NIL)
	map1.Add(MakeKeyword("Host"), MakeString((*res1).Host))
	map1.Add(MakeKeyword("Path"), MakeString((*res1).Path))
	map1.Add(MakeKeyword("RawPath"), MakeString((*res1).RawPath))
	map1.Add(MakeKeyword("ForceQuery"), MakeBool((*res1).ForceQuery))
	map1.Add(MakeKeyword("RawQuery"), MakeString((*res1).RawQuery))
	map1.Add(MakeKeyword("Fragment"), MakeString((*res1).Fragment))
	res = res.Conjoin(map1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

func pathUnescape(s string) Object {
	res1, res2 := url.PathUnescape(s)
	res := EmptyVector
	res = res.Conjoin(MakeString(res1))
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

func queryUnescape(s string) Object {
	res1, res2 := url.QueryUnescape(s)
	res := EmptyVector
	res = res.Conjoin(MakeString(res1))
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

// func user(username string) Object {
// 	return url.User(username)
// 	ABEND124(no public information returned)
// }

// func userPassword(username string, password string) Object {
// 	return url.UserPassword(username, password)
// 	ABEND124(no public information returned)
// }
