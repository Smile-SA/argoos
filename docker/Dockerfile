FROM scratch

MAINTAINER Patrice FERLET <patrice.ferlet@smile.fr>

EXPOSE 3000

ENTRYPOINT ["/argoos"]

LABEL name="Argoos"    \
      version=$VERSION \
      description="Argoos can be used as Docker \
      webhook receiver to update Kubernetes deployment" \
      doc="docker run --rm smilelab/argoos -help"

ADD argoos /argoos
