package ch.epfl.prifiproxy.activities;

import android.os.Bundle;
import android.support.annotation.NonNull;
import android.support.design.widget.FloatingActionButton;
import android.support.design.widget.Snackbar;
import android.support.v7.app.AppCompatActivity;
import android.support.v7.widget.DividerItemDecoration;
import android.support.v7.widget.LinearLayoutManager;
import android.support.v7.widget.RecyclerView;
import android.support.v7.widget.Toolbar;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.TextView;

import java.util.ArrayList;
import java.util.List;

import ch.epfl.prifiproxy.R;
import eu.faircode.netguard.Util;

public class AboutActivity extends AppCompatActivity {

    private RecyclerView aboutList;
    private AboutAdapter adapter;
    private ArrayList<Object> about;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_about);
        Toolbar toolbar = (Toolbar) findViewById(R.id.toolbar);
        setSupportActionBar(toolbar);
        getSupportActionBar().setDisplayHomeAsUpEnabled(true);

        aboutList = findViewById(R.id.aboutList);
        aboutList.setHasFixedSize(true);

        about = new ArrayList<>();

        RecyclerView.ItemDecoration decoration = new DividerItemDecoration(this, DividerItemDecoration.VERTICAL);
        aboutList.addItemDecoration(decoration);

//        about.add(Header(R.string.about_information));
        about.add(Item(R.string.about_app_version, String.valueOf(Util.getSelfVersionCode(this))));
        about.add(Item(R.string.about_protocol_version, "0.1"));

//        about.add(Header(R.string.about_help));
        about.add(Item(R.string.about_support, getString(R.string.about_support_url)));

        LinearLayoutManager linearLayoutManager = new LinearLayoutManager(this);
        aboutList.setLayoutManager(linearLayoutManager);
        adapter = new AboutAdapter(about);
        aboutList.setAdapter(adapter);
    }

    AboutHeader Header(int resId) {
        return new AboutHeader(getString(resId));
    }

    AboutItem Item(int resId, String subtitle) {
        return new AboutItem(getString(resId), subtitle);
    }

    static class AboutHeader {
        String title;

        AboutHeader(String title) {
            this.title = title;
        }
    }

    static class AboutItem {
        String title;
        String subtitle;

        AboutItem(String title, String subtitle) {
            this.title = title;
            this.subtitle = subtitle;
        }
    }

    static class AboutAdapter extends RecyclerView.Adapter<RecyclerView.ViewHolder> {
        private int HEADER = 0;
        private int ITEM = 1;

        private final List<Object> dataset;

        public AboutAdapter(List<Object> dataset) {
            this.dataset = dataset;
        }

        @NonNull
        @Override
        public RecyclerView.ViewHolder onCreateViewHolder(@NonNull ViewGroup parent, int viewType) {
            if (viewType == HEADER) {
                View v = LayoutInflater.from(parent.getContext())
                        .inflate(R.layout.about_list_header, parent, false);

                return new HeaderViewHolder(v);
            } else {
                View v = LayoutInflater.from(parent.getContext())
                        .inflate(R.layout.about_list_item, parent, false);

                return new ItemViewHolder(v);
            }
        }

        @Override
        public void onBindViewHolder(@NonNull RecyclerView.ViewHolder holder, int position) {
            if (holder instanceof HeaderViewHolder) {
                ((HeaderViewHolder) holder).bind((AboutHeader) dataset.get(position));
            } else {
                ((ItemViewHolder) holder).bind((AboutItem) dataset.get(position));
            }
        }

        @Override
        public int getItemCount() {
            return dataset.size();
        }

        @Override
        public int getItemViewType(int position) {
            if (dataset.get(position) instanceof AboutHeader) {
                return HEADER;
            }
            return ITEM;
        }

        static class ItemViewHolder extends RecyclerView.ViewHolder {
            TextView itemText;
            TextView subtitleText;

            ItemViewHolder(View itemView) {
                super(itemView);
                itemText = itemView.findViewById(R.id.title);
                subtitleText = itemView.findViewById(R.id.subtitle);
            }

            void bind(AboutItem aboutItem) {
                itemText.setText(aboutItem.title);
                subtitleText.setText(aboutItem.subtitle);
            }
        }

        static class HeaderViewHolder extends RecyclerView.ViewHolder {
            TextView headerText;

            HeaderViewHolder(View itemView) {
                super(itemView);
                headerText = itemView.findViewById(R.id.title);
            }

            void bind(AboutHeader aboutHeader) {
                headerText.setText(aboutHeader.title);
            }
        }
    }

}
