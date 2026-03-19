FROM postgres:16-alpine
COPY init-db.sh /init-db.sh
RUN chmod +x /init-db.sh
ENTRYPOINT ["/init-db.sh"]
