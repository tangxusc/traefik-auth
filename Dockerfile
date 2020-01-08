FROM scratch
COPY main /
EXPOSE 9999
CMD ["/main"]