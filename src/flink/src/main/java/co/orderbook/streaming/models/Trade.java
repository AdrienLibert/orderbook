import java.io.Serializable;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.flink.api.common.serialization.DeserializationSchema;
import org.apache.flink.api.common.typeinfo.TypeInformation;

public class Trade implements Serializable {
    private String order_id;
    private double quantity;
    private double price;
    private String action;
    private String status;

    // Default constructor
    public Trade() {}

    // Getters and setters
    public String getOrder_id() { return order_id; }
    public void setOrder_id(String order_id) { this.order_id = order_id; }
    public double getQuantity() { return quantity; }
    public void setQuantity(double quantity) { this.quantity = quantity; }
    public double getPrice() { return price; }
    public void setPrice(double price) { this.price = price; }
    public String getAction() { return action; }
    public void setAction(String action) { this.action = action; }
    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }
}


public class TradeDeserializationSchema implements DeserializationSchema<Trade> {
    private static final ObjectMapper objectMapper = new ObjectMapper();

    @Override
    public Trade deserialize(byte[] message) throws IOException {
        return objectMapper.readValue(message, Trade.class);
    }

    @Override
    public boolean isEndOfStream(Trade nextElement) {
        return false;
    }

    @Override
    public TypeInformation<Trade> getProducedType() {
        return TypeInformation.of(Trade.class);
    }
}
