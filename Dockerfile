FROM alpine:latest

# # Create directories
WORKDIR /mossh
# # Expose data volume
VOLUME /mossh

# install mods from latest github binary
RUN apk --no-cache add curl tar
RUN curl -sLO $(curl -s https://api.github.com/repos/charmbracelet/mods/releases/latest | grep "https.*aarch64.apk" | awk '{print $2}' | sed 's|[\"\,]*||g')
RUN apk add --allow-untrusted $(ls | grep -E '\.apk$')

# workaround to prevent slowness in docker when running with a tty
ENV CI "1"

# Expose ports
# SSH
EXPOSE 23234/tcp

# Set the default command
ENTRYPOINT [ "/usr/local/bin/mossh" ]

RUN apk update && apk add --update bash openssh --no-cache

COPY mossh /usr/local/bin/mossh