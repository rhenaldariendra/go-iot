FROM alpine:latest
# Install timezone data (required for Alpine)
RUN apk add --no-cache tzdata

# Set the timezone to Asia/Jakarta
ENV TZ=Asia/Jakarta
RUN ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone \

# Create the /app directory
RUN mkdir -p /app

# Set the working directory to /app
WORKDIR /app

# Copy the application binary
COPY coreApp /app/

# Copy the entire config directory
COPY config /app/config/

# Copy assets file
COPY assets /app/assets/

# Set the default command
CMD [ "/app/coreApp" ]