package com.mogujie.shadow.monitor.furion;

import java.util.HashMap;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.mogujie.shadow.monitor.utils.ConfUtils;
import com.mogujie.shadow.monitor.utils.Constans;
import com.mogujie.tesla.furion.client.FurionClient;
import com.mogujie.tesla.furion.client.FurionClientFactory;

/**
 * @author ziyuan
 * @date 2/5/15 9:45 AM
 * @project shadow-monitor
 */
public class ShadowFurionMock {
    private static final Logger logger = LoggerFactory.getLogger(ShadowFurionMock.class);
    private static FurionClient furionClient = FurionClientFactory.getFurionClient();

    private static HashMap<String, ShadowFurionListener> listeners = new HashMap<String, ShadowFurionListener>();

    public void startListen() {
        List<String> appNames = ConfUtils.readFromConf();

        for (String appname : appNames) {
            String key = appname + Constans.SHADOW_SUFFIX;
            if (listeners.get(key) == null)
                listeners.put(key, new ShadowFurionListener(appname));
            furionClient.addListener(key, listeners.get(key));
        }
    }
}
