# Mutual Funds NAV Backend

<div style="background-color: #d0f0c0; border-radius: 8px; padding: 10px 10px 2px; margin: 4px; text-align: center;">
  <p style="font-weight: bold;">⚠️ Note: This is a learning project and is not intended for any other use. Please use it only for educational purposes.</p>
</div>

## Overview

Welcome to the **Mutual Funds NAV Backend** project! This Go-based backend service provides access to information about Indian mutual funds and their Net Asset Values (NAV). Given a mutual fund ID, you can retrieve the latest NAV details or filter NAV data within a specified date range.

## Features

- **Fetch NAV Information**: Retrieve the NAV details for a specified mutual fund ID.
- **Date Range Filtering**: Optionally filter NAV data for a mutual fund within a specific date range (format: `dd-mm-yyyy`).
- **Database Caching**: The system first checks a remote database for the requested data and fetches from an upstream API only if the data is not available remotely.

## How It Works

1. **Database Lookup**: When a request is made, the service first queries the remote database for the required NAV information.
2. **Upstream API Fetch**: If the data is not found in the remote database, the service fetches it from an upstream API.
3. **Caching**: New data fetched from the upstream API is saved in the remote database for future requests.

## Getting Started

### Prerequisites

- Docker and Docker Compose

### Local Development with Docker

To simplify local development and testing, use Docker. The project includes Docker Compose and a Dockerfile configured for hot-reloading with Air. 

Please update your environment variables in a `.env` file using `.env.example` file before proceeding ahead. 

1. **Build and Start the Docker Containers**

   ```bash
   docker-compose up -d mutual-funds-backend-api
   ```

   This command will build the Docker image and start the application in detached mode. The Air tool will automatically reload the service whenever you make changes to the code.

2. **Access the Application**

   The application should now be running and accessible at `http://localhost:8080`. Adjust your `docker-compose.yml` if you need to change the port.

### Usage

#### API Endpoints

- **Get NAV by Mutual Fund ID**

  ```
  GET /fund/{mutualFundID}
  ```

  **Parameters:**
  - `mutualFundID`: The unique ID for the mutual fund.

- **Get NAV by Mutual Fund ID and Date Range**

  ```
  GET /fund/{mutualFundID}?start={startDate}&end={endDate}
  ```

  **Parameters:**
  - `mutualFundID`: The unique ID for the mutual fund.
  - `startDate`: Start date in `dd-mm-yyyy` format.
  - `endDate`: End date in `dd-mm-yyyy` format.

## Work in Progress

Please note that this project is still under development. While the core functionality is operational, there may be incomplete features and potential bugs. I am actively working on enhancements and improvements.

## Contributing

I welcome contributions to improve this project. If you have suggestions or find issues, please open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact

For questions or further information, please reach out to me or open an issue on the GitHub repository.
