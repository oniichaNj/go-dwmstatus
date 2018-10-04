build:
	go build dwmstatus.go

clean:
	rm dwmstatus

install: build
	cp -vf dwmstatus /usr/local/bin/dwmstatus
