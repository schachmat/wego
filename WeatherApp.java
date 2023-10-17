import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;

public class WeatherApp {

    private static final String API_KEY = "YOUR_API_KEY";
    private static final String API_URL = "http://api.openweathermap.org/data/2.5/weather?q=";
    
    public static void main(String[] args) {
        try {
            BufferedReader reader = new BufferedReader(new InputStreamReader(System.in));
            System.out.print("Enter city name: ");
            String city = reader.readLine();
            
            String apiUrl = API_URL + city + "&appid=" + API_KEY;
            URL url = new URL(apiUrl);
            
            HttpURLConnection connection = (HttpURLConnection) url.openConnection();
            connection.setRequestMethod("GET");
            
            int responseCode = connection.getResponseCode();
            
            if (responseCode == HttpURLConnection.HTTP_OK) {
                BufferedReader in = new BufferedReader(new InputStreamReader(connection.getInputStream()));
                String inputLine;
                StringBuilder response = new StringBuilder();
                
                while ((inputLine = in.readLine()) != null) {
                    response.append(inputLine);
                }
                in.close();
                
                // Parse the JSON response here and display weather information to the user
                System.out.println("Weather data: " + response.toString());
            } else {
                System.out.println("Error: Unable to fetch weather data.");
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
