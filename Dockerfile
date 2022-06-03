FROM library/golang

# Godep for vendoring
RUN go install github.com/tools/godep

# Recompile the standard library without CGO
RUN CGO_ENABLED=0 go install -a std

ENV APP_DIR $GOPATH/home/fish/Foreman-Gateway
RUN mkdir -p $APP_DIR

# Set the entrypoint
ENTRYPOINT (cd $APP_DIR && ./Foreman-Gateway)
ADD . $APP_DIR

# Compile the binary and statically link
RUN cd $APP_DIR && CGO_ENABLED=0 go build -ldflags '-d -w -s'

EXPOSE 8080
EXPOSE 8088
