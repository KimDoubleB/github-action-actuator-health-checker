FROM scratch
EXPOSE 8080
ENTRYPOINT ["/jx-go-lang"]
COPY ./bin/ /