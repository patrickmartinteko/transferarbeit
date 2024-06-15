# Es wird als Basis das offizieller golang Docker Image von Dockerhub verwendet. 
# Es wird spezifisch die Version 1.22.3 und die Variante alpine verwendet.
# Die alpine Variante hat den Vorteil besonders klein zu sein.
FROM golang:1.22.3-alpine

# Wir definieren das Arbeitsverzeichnis. 
WORKDIR /usr/src/app

# Wir kopieren alles aus dem aktuellen Ordner.
COPY . .

# Wir erstellen die Applikation.
RUN go build -v -o /usr/local/bin/app ./...

# Wenn der Container gestartet wird, wird standartmässig die Applikation ausgeführt.
CMD ["app"]