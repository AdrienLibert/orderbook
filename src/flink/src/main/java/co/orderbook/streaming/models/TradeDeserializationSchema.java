package co.orderbook.streaming.models;

import java.io.*;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.flink.api.common.serialization.DeserializationSchema;
import org.apache.flink.api.common.typeinfo.TypeInformation;

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
