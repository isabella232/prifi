package ch.epfl.prifiproxy.viewmodel;

import android.app.Application;
import android.arch.lifecycle.AndroidViewModel;
import android.arch.lifecycle.LiveData;
import android.arch.lifecycle.MediatorLiveData;
import android.content.Context;
import android.support.annotation.NonNull;

import java.lang.ref.WeakReference;
import java.util.List;

import ch.epfl.prifiproxy.persistence.entity.Configuration;
import ch.epfl.prifiproxy.persistence.entity.ConfigurationGroup;
import ch.epfl.prifiproxy.repository.ConfigurationGroupRepository;
import ch.epfl.prifiproxy.repository.ConfigurationRepository;

public class MainViewModel extends AndroidViewModel {
    private final ConfigurationRepository configurationRepository;
    private final ConfigurationGroupRepository groupRepository;

    private LiveData<Configuration> activeConfiguration;
    private LiveData<ConfigurationGroup> activeGroup;
    private LiveData<List<Configuration>> configurations;

    public MainViewModel(@NonNull Application application) {
        super(application);
        configurationRepository = ConfigurationRepository.getInstance(application);
        groupRepository = ConfigurationGroupRepository.getInstance(application);
        activeConfiguration = configurationRepository.getActiveLive();
        activeGroup = groupRepository.getActiveLive();
        configurations = groupRepository.getConfigurationsForActiveGroup();
    }

    public void setActive(Configuration configuration) {
        configurationRepository.setActive(configuration);
    }

    public void updateSettings(WeakReference<Context> context) {
        configurationRepository.updateSettings(context);
    }

    public LiveData<Configuration> getActiveConfiguration() {
        return activeConfiguration;
    }

    public LiveData<ConfigurationGroup> getActiveGroup() {
        return activeGroup;
    }

    public LiveData<List<Configuration>> getConfigurations() {
        return configurations;
    }

    public static class ActiveGroupAndConfig {
        private ConfigurationGroup group;
        private Configuration configuration;

        protected ActiveGroupAndConfig(ConfigurationGroup group, Configuration configuration) {
            this.group = group;
            this.configuration = configuration;
        }

        public ConfigurationGroup getGroup() {
            return group;
        }

        public Configuration getConfiguration() {
            return configuration;
        }
    }
}
