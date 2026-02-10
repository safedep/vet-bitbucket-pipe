FROM ghcr.io/safedep/vet:v1.13.1

RUN apt update -y
RUN apt install curl -y

COPY pipe /
COPY LICENSE.txt pipe.yml README.md /

RUN chmod a+x /*.sh

ENTRYPOINT ["/pipe.sh"]
