FROM busybox

MAINTAINER Patrice FERLET <patrice.ferlet@smile.fr>

ARG VERSION
EXPOSE 3000

ADD https://github.com/Smile-SA/argoos/releases/download/${VERSION}/argoos-linux-x86_64 /argoos
ENTRYPOINT ["/argoos"]

LABEL name="Argoos"    \
      version=$VERSION \
      description="Argoos can be used as Docker \
      webhook receiver to update Kubernetes deployment" \
      doc="docker run --rm smilelab/argoos -help"

