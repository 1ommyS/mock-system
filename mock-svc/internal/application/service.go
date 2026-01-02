package application

type Service struct {
	Repos             Repositories
	Tx                TxManager
	Auth              AuthClient
	DSL               DSLRunner
	OutboxMaxAttempts int
}

func New(repos Repositories, tx TxManager, auth AuthClient, dsl DSLRunner, outboxMaxAttempts int) *Service {
	return &Service{
		Repos:             repos,
		Tx:                tx,
		Auth:              auth,
		DSL:               dsl,
		OutboxMaxAttempts: outboxMaxAttempts,
	}
}
