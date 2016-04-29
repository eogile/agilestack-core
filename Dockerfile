FROM centurylink/ca-certs
MAINTAINER EOGILE "agilestack@eogile.com"

ENV name core

ENV workdir /core

VOLUME /files

WORKDIR $workdir
ADD $name $workdir/$name

CMD ["./core"]
