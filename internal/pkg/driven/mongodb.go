package driven

type MongoAdapter interface {
	Disconnect() error
	Open(collection string) error
}
