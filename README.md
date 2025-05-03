# Routing with Mapbox

This repository demonstrates routing functionality using Mapbox.

## Configuration

1. Obtain a [Mapbox API key](https://www.mapbox.com/).
2. Set the API key as an environment variable:
   ```bash
   export MAPBOX_API_KEY=your_mapbox_api_key
   ```

## How to Run

1. Clone the repository:
   ```bash
   git clone https://github.com/user020603/routing-mapbox.git
   cd routing-mapbox
   ```

2. Install dependencies for the Go backend:
   ```bash
   go mod tidy
   ```

3. Start the application:
   ```bash
   go run main.go
   ```

4. Test the API endpoint:

   - **URL**: `http://localhost:8080/distance`
   - **Method**: `POST`
   - **Request Body**:
     ```json
     {
       "from": "Chung cư Season Avenue",
       "to": "22 Ao Sen, Mộ Lao, Hà Đông, Hà Nội"
     }
     ```
   - **Example Response**:
     ```json
     {
       "distance": 1.410828,
       "duration": 4.85605,
       "from_coords": [
         105.786744,
         20.986808
       ],
       "to_coords": [
         105.78992,
         20.981995
       ],
       "geometry_line": "q~a_CcntdSeAzD{AtEzHrKbA_A?}B\\gEvBwIxBwEtDmE|G}G"
     }
     ```
Happy Routing!
