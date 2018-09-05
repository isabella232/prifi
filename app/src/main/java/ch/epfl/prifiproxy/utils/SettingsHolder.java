package ch.epfl.prifiproxy.utils;

import android.content.Context;
import android.content.SharedPreferences;

import ch.epfl.prifiproxy.R;

import static android.content.Context.MODE_PRIVATE;

/**
 * Read and Update SharedPreferences with this helper class
 */
public class SettingsHolder {
    private String prifiRelayAddress;
    private int prifiRelayPort;
    private int prifiRelaySocksPort;
    private boolean prifiOnly;

    public static SettingsHolder load(Context context) {
        SettingsHolder holder = new SettingsHolder();
        SharedPreferences prifiPrefs = getPreferences(context);
        holder.prifiRelayAddress = prifiPrefs.getString(context.getString(R.string.prifi_config_relay_address), "");
        holder.prifiRelayPort = prifiPrefs.getInt(context.getString(R.string.prifi_config_relay_port), 0);
        holder.prifiRelaySocksPort = prifiPrefs.getInt(context.getString(R.string.prifi_config_relay_socks_port), 0);
        holder.prifiOnly = prifiPrefs.getBoolean(context.getString(R.string.prifi_config_prifionly), true);

        return holder;
    }

    public boolean isValid() {
        return NetworkHelper.isValidIpv4Address(prifiRelayAddress) &&
                NetworkHelper.isValidPort(String.valueOf(prifiRelayPort)) &&
                NetworkHelper.isValidPort(String.valueOf(prifiRelaySocksPort));
    }

    public boolean save(Context context) {
        if (!isValid()) {
            return false;
        }

        // If it's the same, don't apply to prevent false notifications
        SettingsHolder current = SettingsHolder.load(context);
        if (this.equals(current)) {
            return true;
        }

        SharedPreferences prefs = getPreferences(context);
        SharedPreferences.Editor editor = prefs.edit();
        editor.putString(context.getString(R.string.prifi_config_relay_address), prifiRelayAddress);
        editor.putInt(context.getString(R.string.prifi_config_relay_port), prifiRelayPort);
        editor.putInt(context.getString(R.string.prifi_config_relay_socks_port), prifiRelaySocksPort);
        editor.putBoolean(context.getString(R.string.prifi_config_prifionly), prifiOnly);
        editor.apply();
        return true;
    }

    /**
     * Get Prifi SharedPreferences
     */
    public static SharedPreferences getPreferences(Context context) {
        return context.getSharedPreferences(context.getString(R.string.prifi_config_shared_preferences), MODE_PRIVATE);
    }

    public String getPrifiRelayAddress() {
        return prifiRelayAddress;
    }

    public void setPrifiRelayAddress(String prifiRelayAddress) {
        this.prifiRelayAddress = prifiRelayAddress;
    }

    public int getPrifiRelayPort() {
        return prifiRelayPort;
    }

    public void setPrifiRelayPort(int prifiRelayPort) {
        this.prifiRelayPort = prifiRelayPort;
    }

    public int getPrifiRelaySocksPort() {
        return prifiRelaySocksPort;
    }

    public void setPrifiRelaySocksPort(int prifiRelaySocksPort) {
        this.prifiRelaySocksPort = prifiRelaySocksPort;
    }

    public boolean isPrifiOnly() {
        return prifiOnly;
    }

    public void setPrifiOnly(boolean prifiOnly) {
        this.prifiOnly = prifiOnly;
    }

    @SuppressWarnings("SimplifiableIfStatement")
    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof SettingsHolder)) return false;

        SettingsHolder that = (SettingsHolder) o;

        if (prifiRelayPort != that.prifiRelayPort) return false;
        if (prifiRelaySocksPort != that.prifiRelaySocksPort) return false;
        if (prifiOnly != that.prifiOnly) return false;
        return prifiRelayAddress != null ? prifiRelayAddress.equals(that.prifiRelayAddress) : that.prifiRelayAddress == null;
    }
}
