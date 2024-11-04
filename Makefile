ifeq ($(DESTDIR),)
    DESTDIR := /usr/local
endif
ifeq ($(PREFIX),)
    PREFIX := $(DESTDIR)
endif

ds360go: main.go
	go build

install: ds360go
	install -Dm 644 LICENSE ${DESTDIR}/share/licenses/ds360go/LICENSE
	install -Dm 755 ds360go $(DESTDIR)/bin/ds360go
	install -Dm 755 ds360go-stop.sh $(DESTDIR)/bin/ds360go-stop.sh
	install -dm 755 $(DESTDIR)/lib/udev/rules.d/
	sed '/RUN+="/ s|$$|$(PREFIX)/bin/ds360go-stop.sh"\n|' 80-ds360go.rules > $(DESTDIR)/lib/udev/rules.d/80-ds360go.rules
	install -dm 755 $(DESTDIR)/lib/systemd/user/
	sed '/ExecStart=/ s|$$|$(PREFIX)/bin/ds360go|' ds360go.service > $(DESTDIR)/lib/systemd/user/ds360go.service

uninstall:
	rm ${DESTDIR}/share/licenses/ds360go/LICENSE
	rm $(DESTDIR)/bin/ds360go
	rm $(DESTDIR)/bin/ds360go-stop.sh
	rm $(DESTDIR)/lib/udev/rules.d/80-ds360go.rules
	rm $(DESTDIR)/lib/systemd/user/ds360go.service

reload:
	./reload.sh
	
	