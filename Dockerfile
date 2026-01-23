FROM ghcr.io/safedep/vet:v1.12.16

# download curl
# build reporter go app and add binary ih /reporter
# as 2 stage docker build

COPY pipe /
COPY LICENSE.txt pipe.yml README.md /

RUN chmod a+x /*.sh

ENTRYPOINT ["/pipe.sh"]
