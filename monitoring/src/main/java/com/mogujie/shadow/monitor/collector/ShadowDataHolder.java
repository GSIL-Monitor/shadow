package com.mogujie.shadow.monitor.collector;

import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;

/**
 * @author ziyuan
 * @date 2/5/15 3:46 PM
 * @project shadow-monitor
 */
public class ShadowDataHolder {
    private static volatile ConcurrentMap<String, String> furionDatas = new ConcurrentHashMap<String, String>();

    private static ShadowDataHolder instance = new ShadowDataHolder();

    private ShadowDataHolder() {
    }

    public static ShadowDataHolder getInstance() {
        return instance;
    }

    public String getFurionData(String appname) {
        return furionDatas.get(appname);
    }

    public void setFurionData(String furionKey, String furionData) {
        furionDatas.put(furionKey, furionData);
    }
}
