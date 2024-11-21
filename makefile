PREFIX = /usr/local
BASH_COMPLETIONS_DIR = $(DESTDIR)$(PREFIX)/share/bash-completion/completions
USER := $(if $(SUDO_USER),$(SUDO_USER),$(USER))

install:
	cp icw /home/$(USER)/bin
	install -d $(BASH_COMPLETIONS_DIR); \
	install -m 644 completions/icw_bashcompletion.sh $(BASH_COMPLETIONS_DIR)/icw; \
