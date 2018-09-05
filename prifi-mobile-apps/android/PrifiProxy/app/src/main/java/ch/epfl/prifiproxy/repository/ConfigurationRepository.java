package ch.epfl.prifiproxy.repository;

import android.app.Application;
import android.arch.lifecycle.LiveData;
import android.content.Context;
import android.os.AsyncTask;
import android.util.Log;

import java.lang.ref.WeakReference;
import java.util.List;

import ch.epfl.prifiproxy.persistence.AppDatabase;
import ch.epfl.prifiproxy.persistence.dao.ConfigurationDao;
import ch.epfl.prifiproxy.persistence.entity.Configuration;
import ch.epfl.prifiproxy.utils.SettingsHolder;

public class ConfigurationRepository {
    private static final String TAG = "CONFIGURATION_REPO";
    private ConfigurationDao configurationDao;
    private static ConfigurationRepository sInstance;

    public static ConfigurationRepository getInstance(Application application) {
        if (sInstance == null) {
            synchronized (ConfigurationRepository.class) {
                if (sInstance == null) {
                    sInstance = new ConfigurationRepository(application);
                }
            }
        }
        return sInstance;
    }

    ConfigurationRepository(Application application) {
        AppDatabase db = AppDatabase.getDatabase(application);
        configurationDao = db.configurationDao();
    }

    public Configuration getActive() {
        return configurationDao.getActive();
    }

    public LiveData<Configuration> getActiveLive() {
        return configurationDao.getActiveLive();
    }

    public LiveData<Configuration> getConfiguration(int configurationId) {
        return configurationDao.get(configurationId);
    }

    public void updateSettings(WeakReference<Context> context) {
        new UpdateSettingsAsyncTask(configurationDao, context).execute();
    }

    public void setActive(Configuration configuration) {
        new UpdateActiveAsyncTask(configurationDao).execute(configuration);
    }

    public LiveData<List<Configuration>> getConfigurations(int groupId) {
        return configurationDao.getForGroup(groupId);
    }

    public void insert(Configuration configuration) {
        new InsertAsyncTask(configurationDao).execute(configuration);
    }

    public void update(Configuration configuration) {
        new UpdateAsyncTask(configurationDao).execute(configuration);
    }

    public void update(List<Configuration> configurations) {
        new UpdateAsyncTask(configurationDao)
                .execute(configurations.toArray(new Configuration[configurations.size()]));
    }

    public void delete(Configuration... configuration) {
        new DeleteAsyncTask(configurationDao).execute(configuration);
    }

    private static class InsertAsyncTask extends AsyncTask<Configuration, Void, Void> {
        private final ConfigurationDao dao;

        InsertAsyncTask(ConfigurationDao dao) {
            this.dao = dao;
        }

        @Override
        protected Void doInBackground(final Configuration... configurations) {
            if (configurations.length != 1)
                throw new IllegalArgumentException("Must insert one item at a time");

            int count = dao.countConfigurationsForGroups(configurations[0].getGroupId());
            configurations[0].setPriority(count + 1);

            dao.insert(configurations);
            return null;
        }
    }

    private static class UpdateAsyncTask extends AsyncTask<Configuration, Void, Void> {
        private final ConfigurationDao dao;

        UpdateAsyncTask(ConfigurationDao dao) {
            this.dao = dao;
        }

        @Override
        protected Void doInBackground(Configuration... configurations) {
            dao.update(configurations);
            return null;
        }
    }

    //TODO: Improve this
    private static class UpdateSettingsAsyncTask extends AsyncTask<Void, Void, Void> {
        private final ConfigurationDao dao;
        private final WeakReference<Context> context;

        UpdateSettingsAsyncTask(ConfigurationDao dao, WeakReference<Context> context) {
            this.dao = dao;
            this.context = context;
        }

        @Override
        protected Void doInBackground(Void... voids) {
            Configuration config = dao.getActive();
            if (config == null) return null;

            Log.i(TAG, "Updating settings with " + config.getName());

            Context ctx = context.get();
            SettingsHolder settings = SettingsHolder.load(ctx);
            settings.setPrifiRelayAddress(config.getHost());
            settings.setPrifiRelayPort(config.getRelayPort());
            settings.setPrifiRelaySocksPort(config.getSocksPort());
            settings.save(ctx);
            return null;
        }
    }

    private static class UpdateActiveAsyncTask extends AsyncTask<Configuration, Void, Void> {
        private final ConfigurationDao dao;

        UpdateActiveAsyncTask(ConfigurationDao dao) {
            this.dao = dao;
        }

        @Override
        protected Void doInBackground(Configuration... configurations) {
            Configuration newActive = configurations[0];

            Configuration oldActive = dao.getActive();
            if (oldActive != null) {
                oldActive.setActive(false);
                dao.update(oldActive);
            }
            newActive.setActive(true);
            dao.update(newActive);

            return null;
        }
    }

    private static class DeleteAsyncTask extends AsyncTask<Configuration, Void, Void> {
        private final ConfigurationDao dao;

        DeleteAsyncTask(ConfigurationDao dao) {
            this.dao = dao;
        }

        @Override
        protected Void doInBackground(Configuration... configurations) {
            dao.delete(configurations);
            return null;
        }
    }
}
