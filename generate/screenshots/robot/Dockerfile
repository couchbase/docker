FROM playwright/chrome:playwright-1.42.1

# Uncomment this to show playwright debug output
# ENV DEBUG=pw:api

COPY --chown=pwuser:pwuser src /src
COPY entrypoint.sh /

WORKDIR /src

CMD [ "/entrypoint.sh" ]
