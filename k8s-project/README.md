# k8s-project Documentation

## Project Overview

This project is designed to be deployed on Kubernetes and consists of a Python application with a modular structure. The application is built to facilitate easy development and deployment using Docker and Kubernetes.

## Project Structure

```
k8s-project
├── src
│   ├── main.py          # Main entry point of the application
│   └── utils.py         # Utility functions for common tasks
├── k8s
│   ├── deployment.yaml   # Kubernetes deployment configuration
│   ├── service.yaml      # Kubernetes service configuration
│   └── dev-pod.yaml      # Development pod configuration for local mounting
├── scripts
│   └── build-image.sh    # Script to build Docker image
├── Dockerfile             # Instructions for building the Docker image
├── requirements.txt       # Python dependencies
└── README.md              # Project documentation
```

## Setup Instructions

1. **Clone the Repository**

   ```
   git clone <repository-url>
   cd k8s-project
   ```

2. **Install Dependencies**
   Ensure you have Python and pip installed, then run:

   ```
   pip install -r requirements.txt
   ```

3. **Build the Docker Image**
   Use the provided script to build the Docker image:

   ```
   ./scripts/build-image.sh
   ```

4. **Deploy to Kubernetes**
   - Apply the deployment configuration:
     ```
     kubectl apply -f k8s/deployment.yaml
     ```
   - Apply the service configuration:
     ```
     kubectl apply -f k8s/service.yaml
     ```

5. **Development Setup**
   For development purposes, you can use the `dev-pod.yaml` to mount your local directory into a Kubernetes pod:
   ```
   kubectl apply -f k8s/dev-pod.yaml
   ```

## Usage

After deploying the application, you can access it through the service defined in `k8s/service.yaml`. Refer to the service configuration for the appropriate endpoint.

## Contributing

Contributions are welcome! Please submit a pull request or open an issue for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
