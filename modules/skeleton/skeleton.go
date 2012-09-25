// Skeleton contains all the boilerplate you will ever need in any plugin.
package skeleton

type H struct {
	uni *context.Uni
}
func Hooks(uni *context.Uni) *H {
	return &H{uni}
}

type A struct {
	uni *context.Uni
}

func Actions(uni *context.Uni) *A {
	return &A{uni}
}

type V struct {
	uni *context.Uni
}

func Views(uni *context.Uni) *V {
	return &V{uni}
}