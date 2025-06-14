package interfaces

type ListProvider interface {
	QueryList(groupCode int64, site string, msgContext ...any) []byte
	RunIList(msgContext any, groupCode int64, site string) []byte
}

// 新增工厂函数接口实现
var listProviderFactory ListProviderFactory

type ListProviderFactory interface {
	NewListProvider() ListProvider
}

// 注册工厂函数
func RegisterListProvider(factory ListProviderFactory) {
	listProviderFactory = factory
}

// 新增获取实例方法
func NewListProvider() ListProvider {
	if listProviderFactory == nil {
		panic("list provider factory not registered")
	}
	return listProviderFactory.NewListProvider()
}
