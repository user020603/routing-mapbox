const MAPBOX_TOKEN =
  "pk.eyJ1IjoiYWNhbnRob3BoaXMiLCJhIjoiY21hOGNpYWk2MWFyZTJscTFtdndkbzltbiJ9.ci964EVxKJq-2JcQ8Cmlqw";

// URL API backend
const API_URL = "http://localhost:8080";

// Khởi tạo bản đồ
mapboxgl.accessToken = MAPBOX_TOKEN;
const map = new mapboxgl.Map({
  container: "map",
  style: "mapbox://styles/mapbox/streets-v12",
  center: [105.8342, 21.0278], // Mặc định hiển thị Hà Nội
  zoom: 12,
});

// Thêm controls cho bản đồ
map.addControl(new mapboxgl.NavigationControl(), "top-right");
map.addControl(new mapboxgl.FullscreenControl(), "top-right");

// Tạo markers và nguồn dữ liệu tuyến đường
let startMarker, endMarker;

// Đợi bản đồ load xong trước khi thêm các layers
map.on("load", () => {
  // Thêm nguồn dữ liệu cho tuyến đường
  map.addSource("route", {
    type: "geojson",
    data: {
      type: "Feature",
      properties: {},
      geometry: {
        type: "LineString",
        coordinates: [],
      },
    },
  });

  // Thêm layer cho tuyến đường
  map.addLayer({
    id: "route",
    type: "line",
    source: "route",
    layout: {
      "line-join": "round",
      "line-cap": "round",
    },
    paint: {
      "line-color": "#4285f4",
      "line-width": 5,
      "line-opacity": 0.75,
    },
  });
});

// Xử lý form submit
document.getElementById("route-form").addEventListener("submit", async (e) => {
  e.preventDefault();

  const fromAddress = document.getElementById("from-address").value;
  const toAddress = document.getElementById("to-address").value;

  if (!fromAddress || !toAddress) {
    alert("Vui lòng nhập cả địa chỉ xuất phát và địa chỉ đến!");
    return;
  }

  try {
    showLoader(true);
    const routeData = await getRoute(fromAddress, toAddress);
    showLoader(false);
    displayRouteInfo(routeData);
    displayRouteOnMap(routeData);
  } catch (error) {
    showLoader(false);
    alert("Có lỗi xảy ra: " + error.message);
    console.error("Error:", error);
  }
});

// Gọi API để lấy thông tin tuyến đường
async function getRoute(fromAddress, toAddress) {
  const response = await fetch(`${API_URL}/distance`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ from: fromAddress, to: toAddress }),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Server error: ${response.status} - ${errorText}`);
  }

  return await response.json();
}

// Hiển thị thông tin tuyến đường
function displayRouteInfo(routeData) {
  document.getElementById("distance").textContent =
    routeData.distance.toFixed(2);
  document.getElementById("duration").textContent =
    routeData.duration.toFixed(1);
  document.getElementById("route-info").style.display = "block";
}

// Hiển thị tuyến đường trên bản đồ
function displayRouteOnMap(routeData) {
  const { from_coords, to_coords, geometry_line } = routeData;

  // Giải mã polyline thành tọa độ
  const decodedCoordinates = polyline
    .decode(geometry_line)
    .map((point) => [point[1], point[0]]);

  // Cập nhật tuyến đường
  map.getSource("route").setData({
    type: "Feature",
    properties: {},
    geometry: {
      type: "LineString",
      coordinates: decodedCoordinates,
    },
  });

  // Xóa markers cũ nếu có
  if (startMarker) startMarker.remove();
  if (endMarker) endMarker.remove();

  // Tạo markers mới
  startMarker = new mapboxgl.Marker({ color: "#4285F4" })
    .setLngLat(from_coords)
    .setPopup(new mapboxgl.Popup().setHTML("<h6>Điểm xuất phát</h6>"))
    .addTo(map);

  endMarker = new mapboxgl.Marker({ color: "#EA4335" })
    .setLngLat(to_coords)
    .setPopup(new mapboxgl.Popup().setHTML("<h6>Điểm đến</h6>"))
    .addTo(map);

  // Điều chỉnh bản đồ để hiển thị toàn bộ tuyến đường
  const bounds = new mapboxgl.LngLatBounds()
    .extend(from_coords)
    .extend(to_coords);

  // Mở rộng bounds nếu cần để bao gồm toàn bộ tuyến đường
  decodedCoordinates.forEach((coord) => bounds.extend(coord));

  map.fitBounds(bounds, {
    padding: 80,
    maxZoom: 15,
    duration: 1000,
  });
}

// Hiển thị hoặc ẩn loader
function showLoader(show) {
  document.getElementById("loader").style.display = show ? "block" : "none";
}
