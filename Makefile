.PHONY: build run clean update

NAME=webapp_bot
BUILD=docker build -t $(NAME) .
RUN=docker run -d -p 9443:443 --volume $(HOME)/ssl:/app/ssl --name $(NAME) --log-opt max-size=10m --restart=always $(ARGS) $(NAME)
RM=docker rm -f $(NAME)
RMI=docker rmi $(NAME)
PRUNE=docker image prune -f

build:
	$(BUILD)
	$(PRUNE)

run:
	$(RUN)

clean:
	$(RM)
	$(RMI)

update:
	$(BUILD)
	$(RM)
	$(RUN)
	$(PRUNE)