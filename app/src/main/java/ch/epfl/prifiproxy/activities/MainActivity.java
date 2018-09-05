package ch.epfl.prifiproxy.activities;

import android.app.AlertDialog;
import android.app.ProgressDialog;
import android.arch.lifecycle.ViewModelProviders;
import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.content.IntentFilter;
import android.content.SharedPreferences;
import android.content.res.ColorStateList;
import android.content.res.Resources;
import android.graphics.drawable.Drawable;
import android.net.VpnService;
import android.os.AsyncTask;
import android.os.Build;
import android.os.Bundle;
import android.support.annotation.NonNull;
import android.support.design.widget.FloatingActionButton;
import android.support.design.widget.NavigationView;
import android.support.v4.view.GravityCompat;
import android.support.v4.widget.DrawerLayout;
import android.support.v7.app.ActionBarDrawerToggle;
import android.support.v7.app.AppCompatActivity;
import android.support.v7.widget.AppCompatButton;
import android.support.v7.widget.SwitchCompat;
import android.support.v7.widget.Toolbar;
import android.util.Log;
import android.view.MenuItem;
import android.view.View;
import android.widget.ImageView;
import android.widget.TextView;
import android.widget.Toast;

import java.lang.ref.WeakReference;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.atomic.AtomicBoolean;

import ch.epfl.prifiproxy.PrifiProxy;
import ch.epfl.prifiproxy.R;
import ch.epfl.prifiproxy.persistence.entity.Configuration;
import ch.epfl.prifiproxy.persistence.entity.ConfigurationGroup;
import ch.epfl.prifiproxy.services.PrifiService;
import ch.epfl.prifiproxy.ui.MainDrawerRouter;
import ch.epfl.prifiproxy.utils.NetworkHelper;
import ch.epfl.prifiproxy.utils.SettingsHolder;
import ch.epfl.prifiproxy.utils.SystemHelper;
import ch.epfl.prifiproxy.viewmodel.MainViewModel;
import eu.faircode.netguard.ServiceSinkhole;
import eu.faircode.netguard.Util;
import prifiMobile.PrifiMobile;

import static ch.epfl.prifiproxy.utils.Formatting.humanBytes;

public class MainActivity extends AppCompatActivity implements NavigationView.OnNavigationItemSelectedListener {
    private static final String TAG = "PRIFI_MAIN";
    private static final int REQUEST_VPN = 100;

    public static final String EXTRA_APPROVE = "Approve";

    private AtomicBoolean isPrifiServiceRunning;

    private ProgressDialog mProgessDialog;
    private DrawerLayout drawer;

    private BroadcastReceiver mBroadcastReceiver;
    private FloatingActionButton powerButton;
    private TextView textStatus;
    private TextView textVerbose;
    private MainDrawerRouter drawerRouter;
    private MainViewModel viewModel;
    private ConfigurationGroup activeGroup;
    private Configuration activeConfiguration;
    private List<Configuration> configurationList;
    private AlertDialog.Builder dialogBuilder;
    private AppCompatButton configurationButton;

    private Configuration selectedConfiguration;
    private SwitchCompat prifiOnlySwitch;
    private ImageView lockImage;
    private View buttonOverlay;
    private TextView textBytes;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        Toolbar toolbar = findViewById(R.id.toolbar);
        setSupportActionBar(toolbar);

        // Buttons
        powerButton = findViewById(R.id.powerButton);
        configurationButton = findViewById(R.id.configurationButton);
        lockImage = findViewById(R.id.lockImage);
        buttonOverlay = findViewById(R.id.buttonOverlay);

        // Text
        textStatus = findViewById(R.id.textStatus);
        textBytes = findViewById(R.id.textBytes);
        textVerbose = findViewById(R.id.textVerbose);

        // Drawer
        drawer = findViewById(R.id.drawer_layout);
        ActionBarDrawerToggle toggle = new ActionBarDrawerToggle(this, drawer, toolbar
                , R.string.navigation_drawer_open, R.string.navigation_drawer_close);

        drawer.addDrawerListener(toggle);
        toggle.syncState();

        NavigationView navigationView = findViewById(R.id.nav_view);
        drawerRouter = new MainDrawerRouter();
        drawerRouter.addMenu(navigationView);
        navigationView.setNavigationItemSelectedListener(this);

        prifiOnlySwitch = (SwitchCompat) navigationView.getMenu().findItem(R.id.nav_prifionly).getActionView();
        SettingsHolder settings = SettingsHolder.load(this);
        prifiOnlySwitch.setChecked(settings.isPrifiOnly());

        prifiOnlySwitch.setOnCheckedChangeListener((buttonView, isChecked) -> {
            SharedPreferences prefs = SettingsHolder.getPreferences(this);
            prefs.edit()
                    .putBoolean(getString(R.string.prifi_config_prifionly), isChecked)
                    .apply();
        });

        // Actions
        mBroadcastReceiver = new BroadcastReceiver() {
            @Override
            public void onReceive(Context context, Intent intent) {
                String action = intent.getAction();

                if (action != null) {
                    switch (action) {
                        case PrifiService.PRIFI_STOPPED_BROADCAST_ACTION: // Update UI after shutting down PriFi
                            if (mProgessDialog.isShowing()) {
                                mProgessDialog.dismiss();
                            }
                            updateUIInputCapability(false);
                            break;

                        default:
                            break;
                    }
                }

            }
        };

        powerButton.setOnClickListener(view -> {
            boolean isRunning = isPrifiServiceRunning.get();
            if (!isRunning) {
                if (activeConfiguration == null) {
                    Toast.makeText(this, "No Configuration active", Toast.LENGTH_SHORT)
                            .show();
                } else {
                    prepareVpn();
                }
            } else {
                stopPrifiService();
            }
        });

        viewModel = ViewModelProviders.of(this).get(MainViewModel.class);
        viewModel.getActiveConfiguration().observe(this, this::configChanged);
        viewModel.getActiveGroup().observe(this, this::groupChanged);
        viewModel.getConfigurations().observe(this, this::configurationsChanged);

        dialogBuilder = new AlertDialog.Builder(this)
                .setTitle(R.string.title_dialog_configuration_choose)
                .setPositiveButton(android.R.string.ok, (dialog, which) -> {
                    if (selectedConfiguration != null)
                        viewModel.setActive(selectedConfiguration);
                })
                .setNegativeButton(android.R.string.cancel, null)
                .setCancelable(true);

        configurationButton.setOnClickListener(v -> {
            if (configurationList.size() == 0) {
                Toast.makeText(this, "No saved configurations!", Toast.LENGTH_SHORT).show();
            } else {
                dialogBuilder.show();
            }
        });
    }

    private void configurationsChanged(List<Configuration> configurations) {
        configurationList = configurations;
        int selected = -1;
        ArrayList<String> items = new ArrayList<>();
        int i = 0;
        for (Configuration configuration : configurations) {
            items.add(configuration.getName());
            if (configuration.isActive()) {
                selected = i;
            }
            i += 1;
        }
        selectedConfiguration = null;

        dialogBuilder.setSingleChoiceItems(items.toArray(new String[items.size()]),
                selected, (dialog, which) -> selectedConfiguration = configurationList.get(which));
    }

    private void groupChanged(ConfigurationGroup group) {
        activeGroup = group;
        updateView();
    }

    private void configChanged(Configuration configuration) {
        activeConfiguration = configuration;
        if (configuration != null) {
            viewModel.updateSettings(new WeakReference<>(this));
        }
        updateView();
    }

    private void updateView() {
        if (PrifiProxy.isDevFlavor) {
            StringBuilder builder = new StringBuilder();

            if (activeGroup != null) {
                builder.append(activeGroup.getName());
            }

            if (activeConfiguration != null) {
                builder.append("\nConfiguration: ").append(activeConfiguration.getHost()).append(":")
                        .append(activeConfiguration.getRelayPort());
            }
            textVerbose.setText(builder.toString());
        }

        if (activeGroup != null && activeConfiguration != null) {
            configurationButton.setText(getString(R.string.text_button_configuration, activeConfiguration.getName(), activeGroup.getName()));
        } else {
            configurationButton.setText(R.string.text_button_configuration_none);
        }
    }

    @Override
    public void onBackPressed() {
        if (drawer.isDrawerOpen(GravityCompat.START)) {
            drawer.closeDrawer(GravityCompat.START);
        } else {
            super.onBackPressed();
        }
    }

    @Override
    protected void onResume() {
        super.onResume();

        // Check if the PriFi service is running or not
        // Depending on the result, update UI
        isPrifiServiceRunning = new AtomicBoolean(SystemHelper.isMyServiceRunning(PrifiService.class, this));
        updateUIInputCapability(isPrifiServiceRunning.get());

        // Register BroadcastReceiver
        IntentFilter intentFilter = new IntentFilter();
        intentFilter.addAction(PrifiService.PRIFI_STOPPED_BROADCAST_ACTION);
        registerReceiver(mBroadcastReceiver, intentFilter);
    }

    @Override
    protected void onPause() {
        super.onPause();
        unregisterReceiver(mBroadcastReceiver);
    }

    private void prepareVpn() {
        Intent intent = VpnService.prepare(this);
        if (intent == null) {
            Log.i(TAG, "VPN Already Prepared");
            onActivityResult(REQUEST_VPN, RESULT_OK, null);
        } else {
            startActivityForResult(intent, REQUEST_VPN);
        }
    }

    /**
     * Start PriFi "Service" (if not running)
     */
    private void startPrifiService() {
        if (!isPrifiServiceRunning.get()) {
            new StartPrifiAsyncTask(this).execute();
        }
    }

    private void startVpn() {
        ServiceSinkhole.start("UI", this);
    }

    private void stopVpn() {
        ServiceSinkhole.stop("UI", this, false);
    }

    /**
     * Stop PriFi "Core" (if running), the service will be shut down by itself.
     * <p>
     * The stopping process may take 1-2 seconds, so a ProgressDialog has been added to give users some feedback.
     */
    private void stopPrifiService() {
        if (isPrifiServiceRunning.compareAndSet(true, false)) {
            mProgessDialog = ProgressDialog.show(
                    this,
                    getString(R.string.prifi_service_stopping_dialog_title),
                    getString(R.string.prifi_service_stopping_dialog_message)
            );
            stopService(new Intent(this, PrifiService.class));
        }
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        Log.i(TAG, "onActivityResult request=" + requestCode + " result=" + requestCode + " ok=" + (resultCode == RESULT_OK));
        Util.logExtras(data);

        if (requestCode == REQUEST_VPN) {
            // Handle VPN Approval
            if (resultCode == RESULT_OK) {
                startPrifiService();
            } else {
                Toast.makeText(this, getString(R.string.msg_vpn_cancelled), Toast.LENGTH_LONG).show();
            }
        } else {
            Log.w(TAG, "Unknown activity result request=" + requestCode);
            super.onActivityResult(requestCode, resultCode, data);
        }
    }

    private void updateBytesText(long bytes) {
        String s = humanBytes(bytes);
        textBytes.setText(getString(R.string.text_protected_bytes, s));
    }

    /**
     * Depending on the PriFi Service status, we enable or disable some UI elements.
     *
     * @param isServiceRunning Is the PriFi Service running?
     */
    private void updateUIInputCapability(boolean isServiceRunning) {
        int colorId;
        int dp;
        int elevation;
        int lockImageId;
        int overlayVisibility;
        Drawable lockDrawable;

        if (isServiceRunning) {
            colorId = R.color.colorOn;
            dp = 6;
            lockImageId = R.drawable.ic_lock_on;
            overlayVisibility = View.VISIBLE;
            textStatus.setVisibility(View.VISIBLE);
        } else {
            colorId = R.color.colorOff;
            dp = 20;
            lockImageId = R.drawable.ic_lock_off;
            overlayVisibility = View.INVISIBLE;
            textStatus.setVisibility(View.INVISIBLE);
        }

        elevation = (int) (dp * Resources.getSystem().getDisplayMetrics().density);
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
            lockDrawable = getDrawable(lockImageId);
        } else {
            lockDrawable = getResources().getDrawable(lockImageId);
        }

        powerButton.setCompatElevation(elevation);
        powerButton.setBackgroundTintList(ColorStateList.valueOf(getResources().getColor(colorId)));
        lockImage.setImageDrawable(lockDrawable);
        buttonOverlay.setVisibility(overlayVisibility);
        textBytes.setVisibility(overlayVisibility);
        updateBytesText(1024);
    }

    /**
     * An enum that describes the network availability.
     * <p>
     * None: Both PriFi Relay and Socks Server are not available.
     * RELAY_ONLY: Socks Server is not available.
     * SOCKS_ONLY: PriFi Relay is not available.
     * BOTH: Available
     */
    private enum NetworkStatus {
        NONE,
        RELAY_ONLY,
        SOCKS_ONLY,
        BOTH
    }

    /**
     * The Async Task that
     * <p>
     * 1. Checks network availability
     * 2. Starts PriFi Service
     * 3. Updates UI
     */
    private static class StartPrifiAsyncTask extends AsyncTask<Void, Void, NetworkStatus> {

        private final int DEFAULT_PING_TIMEOUT = 3000; // 3s

        // We need this to update UI, but it's a weak reference in order to prevent the memory leak
        private WeakReference<MainActivity> activityReference;

        StartPrifiAsyncTask(MainActivity context) {
            activityReference = new WeakReference<>(context);
        }

        /**
         * Pre Async Execution
         * <p>
         * Show a ProgressDialog, because the network check may take up to 3 seconds.
         */
        @Override
        protected void onPreExecute() {
            MainActivity activity = activityReference.get();

            if (activity != null && !activity.isFinishing()) {
                activity.mProgessDialog = ProgressDialog.show(
                        activity,
                        activity.getString(R.string.check_network_dialog_title),
                        activity.getString(R.string.check_network_dialog_message));
            }
        }

        /**
         * During Async Execution
         * <p>
         * Check the network availability
         *
         * @return relay status: none, relay only, socks only or both
         */
        @Override
        protected NetworkStatus doInBackground(Void... voids) {
            MainActivity activity = activityReference.get();
            if (activity != null && !activity.isFinishing()) {
                SharedPreferences prefs = activity.getSharedPreferences(
                        activity.getString(R.string.prifi_config_shared_preferences), MODE_PRIVATE);

                SettingsHolder settings = SettingsHolder.load(activity);
                try {
                    PrifiMobile.setRelayAddress(settings.getPrifiRelayAddress());
                    PrifiMobile.setRelayPort(settings.getPrifiRelayPort());
                    PrifiMobile.setRelaySocksPort(settings.getPrifiRelaySocksPort());
                    PrifiMobile.setMobileDisconnectWhenNetworkError(!settings.isPrifiOnly());
                } catch (Exception e) {
                    e.printStackTrace();
                }

                boolean isRelayAvailable = NetworkHelper.isHostReachable(
                        settings.getPrifiRelayAddress(),
                        settings.getPrifiRelayPort(),
                        DEFAULT_PING_TIMEOUT);
                boolean isSocksAvailable = NetworkHelper.isHostReachable(
                        settings.getPrifiRelayAddress(),
                        settings.getPrifiRelaySocksPort(),
                        DEFAULT_PING_TIMEOUT);

                if (isRelayAvailable && isSocksAvailable) {
                    return NetworkStatus.BOTH;
                } else if (isRelayAvailable) {
                    return NetworkStatus.RELAY_ONLY;
                } else if (isSocksAvailable) {
                    return NetworkStatus.SOCKS_ONLY;
                } else {
                    return NetworkStatus.NONE;
                }

            } else {
                return NetworkStatus.NONE;
            }
        }

        /**
         * Post Async Execution
         * <p>
         * Start PriFi Service and update UI
         *
         * @param networkStatus relay status
         */
        @Override
        protected void onPostExecute(NetworkStatus networkStatus) {
            MainActivity activity = activityReference.get();

            if (activity != null && !activity.isFinishing()) {
                if (activity.mProgessDialog.isShowing()) {
                    activity.mProgessDialog.dismiss();
                }

                switch (networkStatus) {
                    case NONE:
                        Toast.makeText(activity, activity.getString(R.string.relay_status_none), Toast.LENGTH_LONG).show();
                        break;

                    case RELAY_ONLY:
                        Toast.makeText(activity, activity.getString(R.string.relay_status_relay_only), Toast.LENGTH_LONG).show();
                        break;

                    case SOCKS_ONLY:
                        Toast.makeText(activity, activity.getString(R.string.relay_status_socks_only), Toast.LENGTH_LONG).show();
                        break;

                    case BOTH:
                        activity.isPrifiServiceRunning.set(true);
                        activity.startService(new Intent(activity, PrifiService.class));
                        activity.updateUIInputCapability(true);
                        activity.startVpn();
                        break;

                    default:
                        break;
                }
            }
        }
    }


    @Override
    public boolean onNavigationItemSelected(@NonNull MenuItem item) {
        int id = item.getItemId();
        item.setChecked(true);
        drawer.closeDrawers();

        // If select returns true, the action was performed already
        if (drawerRouter.selected(id, this)) {
            return true;
        }

        Intent intent = null;

        switch (id) {
            case R.id.nav_apps:
                intent = new Intent(this, AppSelectionActivity.class);
                break;
            case R.id.nav_groups:
                intent = new Intent(this, GroupsActivity.class);
                break;
            case R.id.nav_prifionly:
                ((SwitchCompat) item.getActionView()).toggle();
                item.setChecked(false);
                break;
            case R.id.nav_about:
                intent = new Intent(this, AboutActivity.class);
                break;
            default:
                Toast.makeText(this, "Not implemented", Toast.LENGTH_SHORT).show();
        }

        if (intent != null) {
            startActivity(intent);
        }
        return true;
    }
}
