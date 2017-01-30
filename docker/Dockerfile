FROM busybox

MAINTAINER Patrice FERLET <patrice.ferlet@smile.fr>

ENV VERSION 1-beta1


EXPOSE 3000
ADD https://github.com/Smile-SA/argoos/releases/download/v1-beta2/argoos-linux-x86_64 /argoos
ENTRYPOINT ["/argoos"]

LABEL name="Argoos"    \
      version=$VERSION \
      description="Argoos can be used as Docker \
      webhook receiver to update Kubernetes deployment" \
      doc="docker run --rm smilelab/argoos -help"

