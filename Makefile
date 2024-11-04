gantmigration:
	migrate create -dir db/migrations -ext sql $(name)

migrate-down-1:
	migrate -path db/migrations -database postgres://$(username):$(password)@$(host):5432/$(database)?sslmode=disable down 1

migrate-down:
	migrate -path db/migrations -database postgres://$(username):$(password)@$(host):5432/$(database)?sslmode=disable down

migrate:
	migrate -path db/migrations -database postgres://$(username):$(password)@$(host):5432/$(database)?sslmode=disable up

mockgen:
	bash ./bin/generate-mock.sh
