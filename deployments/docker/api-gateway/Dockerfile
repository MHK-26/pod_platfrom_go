# deployments/docker/api-gateway/Dockerfile
FROM nginx:alpine

# Copy nginx config
COPY deployments/docker/api-gateway/nginx.conf /etc/nginx/nginx.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]