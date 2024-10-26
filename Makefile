tidy:
	cd compcont && go mod tidy && cd -
	cd compcont-std/container && go mod tidy && cd -
	cd compcont-std/finder && go mod tidy && cd -
	go work sync