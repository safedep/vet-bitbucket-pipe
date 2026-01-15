FROM ghcr.io/safedep/vet:v1.12.16

COPY pipe /
COPY LICENSE.txt pipe.yml README.md /

RUN chmod a+x /*.sh

ENTRYPOINT ["/pipe.sh"]
