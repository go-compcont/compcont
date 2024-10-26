tidy:
	cd compcont && go mod tidy && cd -
	cd compcont-std && go mod tidy && cd -
	go work sync