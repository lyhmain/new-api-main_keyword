import React, { useState, useEffect, useRef } from 'react';
import {
  Table,
  Button,
  Modal,
  Form,
  Switch,
  Input,
  Toast,
  Popconfirm,
  Tag,
  Space,
  Typography,
  Select,
  DatePicker,
  Card,
} from '@douyinfe/semi-ui';
import { IconRefresh, IconDelete, IconSearch } from '@douyinfe/semi-icons';
import { API, showError } from '../../helpers';

const { Title, Text, Paragraph } = Typography;
const { RangePicker } = DatePicker;

export default function KeywordAudit() {
  const [audits, setAudits] = useState([]);
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState({});
  const [filters, setFilters] = useState({});
  const [pagination, setPagination] = useState({ currentPage: 1, pageSize: 20, total: 0 });
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedAudit, setSelectedAudit] = useState(null);
  const formApiRef = useRef(null);

  const fetchAudits = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        page: pagination.currentPage,
        page_size: pagination.pageSize,
        ...filters,
      });
      const res = await API.get(`/api/keyword-audit?${params}`);
      if (res.data.success) {
        setAudits(res.data.data.items || []);
        setPagination((p) => ({ ...p, total: res.data.data.total }));
      }
    } catch (err) {
      showError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const res = await API.get('/api/keyword-audit/stats?days=7');
      if (res.data.success) {
        setStats(res.data.data || {});
      }
    } catch (err) {
      showError(err.message);
    }
  };

  useEffect(() => {
    fetchAudits();
    fetchStats();
  }, [pagination.currentPage, filters]);

  const handleSearch = () => {
    const values = formApiRef.current.getValues();
    const newFilters = {};
    if (values.keyword) newFilters.keyword = values.keyword;
    if (values.username) newFilters.username = values.username;
    if (values.action) newFilters.action = values.action;
    if (values.processed !== undefined) newFilters.processed = values.processed;
    setFilters(newFilters);
    setPagination((p) => ({ ...p, currentPage: 1 }));
  };

  const handleReset = () => {
    formApiRef.current.reset();
    setFilters({});
    setPagination((p) => ({ ...p, currentPage: 1 }));
  };

  const handleProcess = async (id, note = '') => {
    try {
      const res = await API.post(`/api/keyword-audit/${id}/process`, { note });
      if (res.data.success) {
        Toast.success('标记已处理成功');
        fetchAudits();
        fetchStats();
      }
    } catch (err) {
      showError(err.message);
    }
  };

  const handleDelete = async (id) => {
    try {
      await API.delete(`/api/keyword-audit/${id}`);
      Toast.success('删除成功');
      fetchAudits();
    } catch (err) {
      showError(err.message);
    }
  };

  const handleViewDetail = (record) => {
    setSelectedAudit(record);
    setDetailModalVisible(true);
  };

  const handleDeleteOld = async () => {
    try {
      const res = await API.delete('/api/keyword-audit/old?days=30');
      if (res.data.success) {
        Toast.success(res.data.message);
        fetchAudits();
        fetchStats();
      }
    } catch (err) {
      showError(err.message);
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '关键词', dataIndex: 'keyword', ellipsis: true },
    {
      title: '操作',
      dataIndex: 'action',
      render: (text) => (
        <Tag color={text === 'replace' ? 'blue' : text === 'block' ? 'red' : 'orange'}>
          {text === 'replace' ? '替换' : text === 'block' ? '阻止' : '审计'}
        </Tag>
      ),
    },
    { title: '用户名', dataIndex: 'username', ellipsis: true },
    { title: '模型', dataIndex: 'model', ellipsis: true },
    { title: '请求类型', dataIndex: 'request_type' },
    {
      title: '状态',
      dataIndex: 'processed',
      render: (text) => (
        <Tag color={text ? 'green' : 'red'}>{text ? '已处理' : '待处理'}</Tag>
      ),
    },
    { title: '创建时间', dataIndex: 'created_at', render: (text) => new Date(text).toLocaleString() },
    {
      title: '操作',
      render: (_, record) => (
        <Space>
<Button type="tertiary" size="small" onClick={() => handleViewDetail(record)}>
              详情
            </Button>
          {!record.processed && (
            <Button
              type="secondary"
              size="small"
              onClick={() => handleProcess(record.id)}
            >
              标记已处理
            </Button>
          )}
          <Popconfirm title="确认删除？" onConfirm={() => handleDelete(record.id)}>
            <Button type="danger" icon={<IconDelete />} size="small">
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title heading={4}>关键词审计记录</Title>
        <Space>
          <Popconfirm title="确认删除30天前的已处理记录？" onConfirm={handleDeleteOld}>
            <Button type="danger" icon={<IconDelete />}>
              清理旧记录
            </Button>
          </Popconfirm>
          <Button icon={<IconRefresh />} onClick={() => { fetchAudits(); fetchStats(); }}>
            刷新
          </Button>
        </Space>
      </div>

      <div style={{ marginBottom: 16, display: 'flex', gap: 16 }}>
        <Tag color="blue">总记录: {stats.total || 0}</Tag>
        <Tag color="green">已替换: {stats.replaced || 0}</Tag>
        <Tag color="red">已阻止: {stats.blocked || 0}</Tag>
        <Tag color="orange">待处理: {stats.unprocessed || 0}</Tag>
      </div>

      <Card style={{ marginBottom: 16 }}>
        <Form layout="horizontal" getFormApi={api => formApiRef.current = api} onSubmit={handleSearch}>
          <Space wrap>
            <Form.Input field="keyword" label="关键词" placeholder="搜索关键词" />
            <Form.Input field="username" label="用户名" placeholder="搜索用户名" />
            <Form.Select field="action" label="操作" placeholder="选择操作">
              <Select.Option value="replace">替换</Select.Option>
              <Select.Option value="block">阻止</Select.Option>
              <Select.Option value="audit">审计</Select.Option>
            </Form.Select>
            <Form.Select field="processed" label="状态" placeholder="选择状态">
              <Select.Option value="true">已处理</Select.Option>
              <Select.Option value="false">待处理</Select.Option>
            </Form.Select>
            <Button type="primary" icon={<IconSearch />} onClick={handleSearch}>
              搜索
            </Button>
            <Button onClick={handleReset}>重置</Button>
          </Space>
        </Form>
      </Card>

      <Table
        columns={columns}
        dataSource={audits}
        loading={loading}
        rowKey="id"
        pagination={{
          currentPage: pagination.currentPage,
          pageSize: pagination.pageSize,
          total: pagination.total,
          onPageChange: (page) => setPagination((p) => ({ ...p, currentPage: page })),
        }}
      />

      <Modal
        title="审计记录详情"
        visible={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={
          <Button onClick={() => setDetailModalVisible(false)}>关闭</Button>
        }
        width={600}
      >
        {selectedAudit && (
          <div>
            <p><strong>ID:</strong> {selectedAudit.id}</p>
            <p><strong>关键词:</strong> {selectedAudit.keyword}</p>
            <p><strong>操作:</strong> {selectedAudit.action}</p>
            <p><strong>用户名:</strong> {selectedAudit.username}</p>
            <p><strong>模型:</strong> {selectedAudit.model}</p>
            <p><strong>请求类型:</strong> {selectedAudit.request_type}</p>
            <p><strong>IP地址:</strong> {selectedAudit.ip_address}</p>
            <p><strong>创建时间:</strong> {new Date(selectedAudit.created_at).toLocaleString()}</p>
            <p><strong>上下文内容:</strong></p>
            <Paragraph style={{ background: '#f5f5f5', padding: 12, borderRadius: 4, maxHeight: 300, overflow: 'auto' }}>
              {selectedAudit.context}
            </Paragraph>
          </div>
        )}
      </Modal>
    </div>
  );
}
