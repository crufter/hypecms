// Skeleton contains all the boilerplate you will ever need in any plugin.
package skeleton

// This is required.
// Member must not be *context.Uni.
type H struct {
	uni *context.Uni
}

// This is optional.
// Init param must not be *context.Uni.
func (c *C) Init(uni *context.Uni) {
	c.uni = uni
}