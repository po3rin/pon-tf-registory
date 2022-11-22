clean:
	rm -rf .terraform
	rm .terraform.lock.hcl

insert:
	curl -X POST -H "Content-Type: application/json" -d '@sample_request.json' 'localhost:8080/v1/providers/po3rin/berglas/0.1.1/regist'