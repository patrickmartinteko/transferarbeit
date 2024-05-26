# Use the official Node.js image as the base image
FROM node:22

# Set the working directory inside the container
WORKDIR /usr/src/app

# Copy the package.json and package-lock.json files to the container
COPY package.json ./

# Install the dependencies
RUN npm install express
RUN npm install mongodb
RUN npm install nodemon
RUN npm install ws

# Copy the source code to the container
COPY . .

# Start the server when the container starts
CMD ["node", "app.js"]
