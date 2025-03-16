package co.orderbook.streaming.models;

import java.io.Serializable;

public class Trade implements Serializable {
    private String trade_id;
    private String order_id;
    private double quantity;
    private double price;
    private String action;
    private String status;

    // Default constructor
    public Trade() {}

    // Getters and setters
    public String getTrade_id() { return trade_id; }
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