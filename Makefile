.PHONY: build run clean update

NAME=bot
BUILD=docker build -t $(NAME) .
RUN=docker run -d --name $(NAME) --log-opt max-size=10m --restart=always $(ARGS) $(NAME)
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