package com.mogujie.shadow.monitor.model;

/**
 * @author ziyuan
 * @date 2/4/15 6:27 PM
 * @project shadow-monitor
 */
public class ShadowRedisInfo {

    private String appName;

    private String ip;
    private int port;

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

    private long used_memory;
    private long used_memory_peak; // 没有free_memory
    private double mem_fragmentation_ratio;

    private long total_commands_processed;
    private long total_connections_received;
    private long connected_clients;

    private long keyspace_hits;
    private long keyspace_misses;

    private long expired_keys;
    private long evicted_keys;

    /**
     * @return the appName
     */
    public String getAppName() {
        return appName;
    }

    /**
     * @param appName
     *            the appName to set
     */
    public void setAppName(String appName) {
        this.appName = appName;
    }

    public long getUsed_memory() {
        return used_memory;
    }

    public void setUsed_memory(long used_memory) {
        this.used_memory = used_memory;
    }

    public long getUsed_memory_peak() {
        return used_memory_peak;
    }

    public void setUsed_memory_peak(long used_memory_peak) {
        this.used_memory_peak = used_memory_peak;
    }

    public double getMem_fragmentation_ratio() {
        return mem_fragmentation_ratio;
    }

    public void setMem_fragmentation_ratio(double mem_fragmentation_ratio) {
        this.mem_fragmentation_ratio = mem_fragmentation_ratio;
    }

    public long getTotal_commands_processed() {
        return total_commands_processed;
    }

    public void setTotal_commands_processed(long total_commands_processed) {
        this.total_commands_processed = total_commands_processed;
    }

    public long getTotal_connections_received() {
        return total_connections_received;
    }

    public void setTotal_connections_received(long total_connections_received) {
        this.total_connections_received = total_connections_received;
    }

    public long getConnected_clients() {
        return connected_clients;
    }

    public void setConnected_clients(long connected_clients) {
        this.connected_clients = connected_clients;
    }

    public long getKeyspace_hits() {
        return keyspace_hits;
    }

    public void setKeyspace_hits(long keyspace_hits) {
        this.keyspace_hits = keyspace_hits;
    }

    public long getKeyspace_misses() {
        return keyspace_misses;
    }

    public void setKeyspace_misses(long keyspace_misses) {
        this.keyspace_misses = keyspace_misses;
    }

    public long getExpired_keys() {
        return expired_keys;
    }

    public void setExpired_keys(long expired_keys) {
        this.expired_keys = expired_keys;
    }

    public long getEvicted_keys() {
        return evicted_keys;
    }

    public void setEvicted_keys(long evicted_keys) {
        this.evicted_keys = evicted_keys;
    }

    @Override
    public boolean equals(Object obj) {
        if (this == obj)
            return true;
        if (obj == null)
            return false;
        if (getClass() != obj.getClass())
            return false;
        ShadowRedisInfo other = (ShadowRedisInfo) obj;
        if (appName == null) {
            if (other.appName != null)
                return false;
        } else if (!appName.equals(other.appName))
            return false;
        if (connected_clients != other.connected_clients)
            return false;
        if (evicted_keys != other.evicted_keys)
            return false;
        if (expired_keys != other.expired_keys)
            return false;
        if (ip == null) {
            if (other.ip != null)
                return false;
        } else if (!ip.equals(other.ip))
            return false;
        if (keyspace_hits != other.keyspace_hits)
            return false;
        if (keyspace_misses != other.keyspace_misses)
            return false;
        if (Double.doubleToLongBits(mem_fragmentation_ratio) != Double.doubleToLongBits(other.mem_fragmentation_ratio))
            return false;
        if (port != other.port)
            return false;
        if (total_commands_processed != other.total_commands_processed)
            return false;
        if (total_connections_received != other.total_connections_received)
            return false;
        if (used_memory != other.used_memory)
            return false;
        if (used_memory_peak != other.used_memory_peak)
            return false;
        return true;
    }

    @Override
    public int hashCode() {
        final int prime = 31;
        int result = 1;
        result = prime * result + ((appName == null) ? 0 : appName.hashCode());
        result = prime * result + (int) (connected_clients ^ (connected_clients >>> 32));
        result = prime * result + (int) (evicted_keys ^ (evicted_keys >>> 32));
        result = prime * result + (int) (expired_keys ^ (expired_keys >>> 32));
        result = prime * result + ((ip == null) ? 0 : ip.hashCode());
        result = prime * result + (int) (keyspace_hits ^ (keyspace_hits >>> 32));
        result = prime * result + (int) (keyspace_misses ^ (keyspace_misses >>> 32));
        long temp;
        temp = Double.doubleToLongBits(mem_fragmentation_ratio);
        result = prime * result + (int) (temp ^ (temp >>> 32));
        result = prime * result + port;
        result = prime * result + (int) (total_commands_processed ^ (total_commands_processed >>> 32));
        result = prime * result + (int) (total_connections_received ^ (total_connections_received >>> 32));
        result = prime * result + (int) (used_memory ^ (used_memory >>> 32));
        result = prime * result + (int) (used_memory_peak ^ (used_memory_peak >>> 32));
        return result;
    }

    @Override
    public String toString() {
        return "ShadowRedisInfo [AppName=" + appName + ", ip=" + ip + ", port=" + port + ", used_memory=" + used_memory
                + ", used_memory_peak=" + used_memory_peak + ", mem_fragmentation_ratio=" + mem_fragmentation_ratio
                + ", total_commands_processed=" + total_commands_processed + ", total_connections_received="
                + total_connections_received + ", connected_clients=" + connected_clients + ", keyspace_hits="
                + keyspace_hits + ", keyspace_misses=" + keyspace_misses + ", expired_keys=" + expired_keys
                + ", evicted_keys=" + evicted_keys + "]";
    }

    public String toCustom() {
        // 这边可以变成sentry相关的数据
        return null;
    }
}
