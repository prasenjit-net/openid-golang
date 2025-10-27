import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  Select,
  Button,
  Space,
  Typography,
  message,
  Row,
  Col,
} from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';

const { Title } = Typography;
const { TextArea } = Input;

const UserCreate = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (values: any) => {
    try {
      setSubmitting(true);
      const response = await fetch('/api/admin/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(values),
      });
      if (!response.ok) throw new Error('Failed to create user');
      const data = await response.json();
      message.success('User created successfully');
      navigate(`/users/${data.id}`);
    } catch (error) {
      message.error('Failed to create user');
      console.error('Failed to create user:', error);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/users')}
          >
            Back to Search
          </Button>
          <Title level={2} style={{ margin: 0 }}>Create New User</Title>
        </Space>
      </div>

      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        initialValues={{ role: 'user' }}
      >
        <Card bordered={false} title="Basic Information" style={{ marginBottom: 24 }}>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Username"
                name="username"
                rules={[{ required: true, message: 'Please enter username' }]}
              >
                <Input />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="Role" name="role">
                <Select>
                  <Select.Option value="user">User</Select.Option>
                  <Select.Option value="admin">Admin</Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Email"
                name="email"
                rules={[
                  { required: true, message: 'Please enter email' },
                  { type: 'email', message: 'Please enter a valid email' },
                ]}
              >
                <Input />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Password"
                name="password"
                rules={[{ required: true, message: 'Please enter password' }]}
              >
                <Input.Password />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="Full Name"
            name="name"
            rules={[{ required: true, message: 'Please enter name' }]}
          >
            <Input />
          </Form.Item>
        </Card>

        <Card bordered={false} title="Profile Claims (Optional)" style={{ marginBottom: 24 }}>
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item label="Given Name" name="given_name">
                <Input />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="Family Name" name="family_name">
                <Input />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="Middle Name" name="middle_name">
                <Input />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="Nickname" name="nickname">
                <Input />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="Preferred Username" name="preferred_username">
                <Input />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="Profile URL" name="profile">
                <Input placeholder="https://example.com/profile" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="Picture URL" name="picture">
                <Input placeholder="https://example.com/avatar.jpg" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="Website" name="website">
                <Input placeholder="https://example.com" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="Gender" name="gender">
                <Select allowClear>
                  <Select.Option value="male">Male</Select.Option>
                  <Select.Option value="female">Female</Select.Option>
                  <Select.Option value="other">Other</Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item label="Birthdate" name="birthdate">
                <Input placeholder="YYYY-MM-DD" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="Timezone" name="zoneinfo">
                <Input placeholder="America/New_York" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="Locale" name="locale">
                <Input placeholder="en-US" />
              </Form.Item>
            </Col>
          </Row>
        </Card>

        <Card bordered={false} title="Contact Information (Optional)" style={{ marginBottom: 24 }}>
          <Form.Item label="Phone Number" name="phone_number">
            <Input placeholder="+1-234-567-8900" />
          </Form.Item>
        </Card>

        <Card bordered={false} title="Address (Optional)" style={{ marginBottom: 24 }}>
          <Form.Item label="Formatted Address" name={['address', 'formatted']}>
            <TextArea rows={2} placeholder="Full formatted address" />
          </Form.Item>

          <Form.Item label="Street Address" name={['address', 'street_address']}>
            <Input />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="City/Locality" name={['address', 'locality']}>
                <Input />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="State/Region" name={['address', 'region']}>
                <Input />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="Postal Code" name={['address', 'postal_code']}>
                <Input />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="Country" name={['address', 'country']}>
                <Input />
              </Form.Item>
            </Col>
          </Row>
        </Card>

        <Card bordered={false}>
          <Space>
            <Button
              type="primary"
              htmlType="submit"
              icon={<SaveOutlined />}
              loading={submitting}
            >
              Create User
            </Button>
            <Button onClick={() => navigate('/users')}>
              Cancel
            </Button>
          </Space>
        </Card>
      </Form>
    </>
  );
};

export default UserCreate;
