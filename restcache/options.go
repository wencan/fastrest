package restcache

// Validatable 缓存有效性检查接口。缓存对象可选实现。
// 警告：注意实现中ValidCache方法的接收者，一般应该是结构体对象，而不是结构体指针。
// 失效的缓存对象可能会影响结果，可选实现Resetable接口。
type Validatable interface {
	// IsValidCache 判断缓存对象是否还有效。返回true表示有效。
	// 如果无效，等同没找到缓存存储数据。
	IsValidCache() bool
}

// Resetable 支持重置的对象接口。
type Resetable interface {
	// Reset 重置。
	Reset()
}
