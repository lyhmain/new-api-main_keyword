import React, { useState, useEffect, useRef } from 'react';
import {
  Table,
  Button,
  Modal,
  Form,
  Switch,
  Input,
  InputNumber,
  Toast,
  Popconfirm,
  Tag,
  Upload,
  Space,
  Typography,
} from '@douyinfe/semi-ui';
import { IconUpload, IconDownload, IconEdit, IconDelete, IconPlus } from '@douyinfe/semi-icons';
import { API, showError } from '../../helpers';

const { Title, Text } = Typography;

export default function KeywordReplacement() {
  const [replacements, setReplacements] = useState([]);
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState({});
  const [modalVisible, setModalVisible] = useState(false);
  const [editingItem, setEditingItem] = useState(null);
  const formApiRef = useRef(null);
  const [pagination, setPagination] = useState({ currentPage: 1, pageSize: 20, total: 0 });

  const fetchReplacements = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/keyword-replacement');
      if (res.data.success) {
        setReplacements(res.data.data.items || []);
        setStats(res.data.data.stats || {});
      }
    } catch (err) {
      showError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchReplacements();
  }, []);

  const handleSubmit = async () => {
    try {
      const values = await formApiRef.current.validate();
      if (editingItem) {
        await API.put(`/api/keyword-replacement/${editingItem.id}`, values);
        Toast.success('更新成功');
      } else {
        await API.post('/api/keyword-replacement', values);
        Toast.success('创建成功');
      }
      setModalVisible(false);
      formApiRef.current.reset();
      fetchReplacements();
    } catch (err) {
      showError(err.message);
    }
  };

  const handleDelete = async (id) => {
    try {
      await API.delete(`/api/keyword-replacement/${id}`);
      Toast.success('删除成功');
      fetchReplacements();
    } catch (err) {
      showError(err.message);
    }
  };

  const handleEdit = (record) => {
    setEditingItem(record);
    setModalVisible(true);
  };

  useEffect(() => {
    if (modalVisible && editingItem && formApiRef.current) {
      formApiRef.current.setValues(editingItem);
    }
  }, [modalVisible, editingItem]);

  const handleExport = async () => {
    try {
      window.open(`${API.defaults.baseURL}/api/keyword-replacement/export`, '_blank');
    } catch (err) {
      showError(err.message);
    }
  };

  const handleImport = async (file) => {
    const formData = new FormData();
    formData.append('file', file);
    try {
      const res = await API.post('/api/keyword-replacement/import', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      if (res.data.success) {
        Toast.success(res.data.message);
        fetchReplacements();
      }
    } catch (err) {
      showError(err.message);
    }
    return false;
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '关键词', dataIndex: 'keyword', ellipsis: true },
    { title: '替换词', dataIndex: 'replacement', ellipsis: true },
    {
      title: '启用',
      dataIndex: 'enabled',
      render: (text) => <Tag color={text ? 'green' : 'red'}>{text ? '是' : '否'}</Tag>,
    },
    {
      title: '正则',
      dataIndex: 'is_regex',
      render: (text) => <Tag color={text ? 'blue' : 'grey'}>{text ? '是' : '否'}</Tag>,
    },
    {
      title: '区分大小写',
      dataIndex: 'case_sensitive',
      render: (text) => <Tag color={text ? 'blue' : 'grey'}>{text ? '是' : '否'}</Tag>,
    },
    { title: '优先级', dataIndex: 'priority', width: 80 },
    { title: '描述', dataIndex: 'description', ellipsis: true },
    {
      title: '操作',
      render: (_, record) => (
        <Space>
          <Button type="tertiary" icon={<IconEdit />} size="small" onClick={() => handleEdit(record)}>
            编辑
          </Button>
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
        <Title heading={4}>关键词替换管理</Title>
        <Space>
          <Upload
            action=""
            beforeUpload={handleImport}
            showUploadList={false}
            accept=".csv"
          >
            <Button icon={<IconUpload />}>导入CSV</Button>
          </Upload>
          <Button icon={<IconDownload />} onClick={handleExport}>
            导出CSV
          </Button>
          <Button
            type="primary"
            icon={<IconPlus />}
            onClick={() => {
              setEditingItem(null);
              formApiRef.current?.reset();
              setModalVisible(true);
            }}
          >
            添加
          </Button>
        </Space>
      </div>

      <div style={{ marginBottom: 16, display: 'flex', gap: 16 }}>
        <Tag color="blue">总数: {stats.total || 0}</Tag>
        <Tag color="green">启用: {stats.enabled || 0}</Tag>
        <Tag color="purple">正则: {stats.regex_count || 0}</Tag>
        <Tag color="orange">待审计: {stats.with_audit || 0}</Tag>
      </div>

      <Table
        columns={columns}
        dataSource={replacements}
        loading={loading}
        rowKey="id"
        pagination={{
          currentPage: pagination.currentPage,
          pageSize: pagination.pageSize,
          total: replacements.length,
          onPageChange: (page) => setPagination((p) => ({ ...p, currentPage: page })),
        }}
      />

      <Modal
        title={editingItem ? '编辑关键词' : '添加关键词'}
        visible={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false);
          formApiRef.current?.reset();
        }}
        okText="确定"
        cancelText="取消"
      >
        <Form getFormApi={api => formApiRef.current = api} labelPosition="left" labelWidth={120}>
          <Form.Input
            field="keyword"
            label="关键词"
            rules={[{ required: true, message: '请输入关键词' }]}
            placeholder="支持文本或正则表达式"
          />
          <Form.Input
            field="replacement"
            label="替换词"
            rules={[{ required: true, message: '请输入替换词' }]}
            placeholder="替换后的文本"
          />
          <Form.Switch field="enabled" label="启用" />
          <Form.Switch field="is_regex" label="正则表达式" />
          <Form.Switch field="case_sensitive" label="区分大小写" />
          <Form.InputNumber field="priority" label="优先级" min={0} />
          <Form.Input field="description" label="描述" />
          <Form.InputNumber field="audit_threshold" label="审计阈值" min={0} />
        </Form>
      </Modal>
    </div>
  );
}
