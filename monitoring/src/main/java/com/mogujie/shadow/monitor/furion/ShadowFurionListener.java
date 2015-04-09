package com.mogujie.shadow.monitor.furion;

import com.mogujie.shadow.monitor.collector.ShadowDataHolder;
import com.mogujie.tesla.furion.client.FurionListener;

/**
 * @author ziyuan
 * @date 2/5/15 9:14 AM
 * @project shadow-monitor
 */
public class ShadowFurionListener implements FurionListener {

    private String appName;
    private String config;

    public ShadowFurionListener(String appName) {
        this.appName = appName;
    }

    public String getAppName() {
        return appName;
    }

    public void setAppName(String appName) {
        this.appName = appName;
    }

    public String getConfig() {
        return config;
    }

    public void valueReceived(String value) {
        // 这边拿到的时最新的key值，即APP对应的Ring的名字
        ShadowDataHolder.getInstance().setFurionData(appName, value);
    }
}
