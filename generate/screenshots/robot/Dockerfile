FROM mcr.microsoft.com/playwright:bionic

# Uncomment this to show playwright debug output
# ENV DEBUG=pw:api

COPY src /src
COPY entrypoint.sh /

WORKDIR /src

CMD [ "/entrypoint.sh" ]
