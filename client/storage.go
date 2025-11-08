package client

type Storage interface {
	Set(key string, token Token) error
	Get(key string) (Token, error)
}
