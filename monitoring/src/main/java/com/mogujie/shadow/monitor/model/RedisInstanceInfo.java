package com.mogujie.shadow.monitor.model;

/**
 * @author ziyuan
 * @date 2/5/15 10:21 AM
 * @project shadow-monitor
 */
public class RedisInstanceInfo {

    private String ip;
    private int    port;

    public String getIp() {
        return ip;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }

    public int getPort() {
        return port;
    }

    public void setPort(int port) {
        this.port = port;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof RedisInstanceInfo)) return false;

        RedisInstanceInfo that = (RedisInstanceInfo) o;

        if (port != that.port) return false;
        if (ip != null ? !ip.equals(that.ip) : that.ip != null) return false;

        return true;
    }

    @Override
    public int hashCode() {
        int result = ip != null ? ip.hashCode() : 0;
        result = 31 * result + port;
        return result;
    }

    @Override
    public String toString() {
        return "RedisInstanceInfo{" +
                "ip='" + ip + '\'' +
                ", port=" + port +
                '}';
    }
}
