package ch.epfl.prifiproxy.utils;

import java.text.DecimalFormat;
import java.text.NumberFormat;

public class Formatting {
    private static final String[] PREFIXES = new String[]{"", "K", "M", "G", "T", "P", "E", "Z", "Y"};

    private static double log2(long n) {
        return (Math.log(n) / Math.log(2));
    }

    public static String humanBytes(long bytes) {
        return humanBytes(bytes, 1);
    }

    public static String humanBytes(long bytes, int fractionDigits) {
        long logSize = (long) log2(bytes);
        int index = (int) (logSize / 10); // 2^10 = 1024

        double displaySize = bytes / Math.pow(2, index * 10);
        NumberFormat df = DecimalFormat.getInstance();
        df.setMaximumFractionDigits(fractionDigits);
        return df.format(displaySize) + PREFIXES[index] + "B";
    }
}
