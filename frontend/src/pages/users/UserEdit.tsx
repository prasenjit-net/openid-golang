import { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  Select,
  Button,
  Space,
  Typography,
  Spin,
  Alert,
  message,
  Row,
  Col,
} from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';
import { useUser, useUpdateUser } from '../../hooks/useApi';

const { Title } = Typography;
const { TextArea } = Input;

const UserEdit = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  
  const { data: user, isLoading: loading, error: queryError } = useUser(id || '');
  const updateUserMutation = useUpdateUser();

  useEffect(() => {
    if (user) {
      form.setFieldsValue(user);
    }
  }, [user, form]);

  const handleSubmit = async (values: any) => {
    try {
      await updateUserMutation.mutateAsync({ id: id!, ...values });
      message.success('User updated successfully');
      navigate(`/users/${id}`);
    } catch (error) {
      message.error('Failed to update user');
      console.error('Failed to update user:', error);
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: '50px' }}>
        <Spin size="large" tip="Loading user..." />
      </div>
    );
  }

  if (!user) {
    return (
      <>
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/users')}
          style={{ marginBottom: 16 }}
        >
          Back to Search
        </Button>
        <Alert
          message="Error"
          description={'User not found'}
          type="error"
          showIcon
        />
      </>
    );
  }

  return (
    <>
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate(`/users/${id}`)}
          >
            Back to Details
          </Button>
          <Title level={2} style={{ margin: 0 }}>Edit User</Title>
        </Space>
      </div>

      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        initialValues={user}
      >
        <Card bordered={false} title="Basic Information" style={{ marginBottom: 24 }}>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Username"
                name="username"
                rules={[{ required: true, message: 'Please enter username' }]}
              >
                <Input disabled />
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
                help="Leave blank to keep current password"
              >
                <Input.Password placeholder="Leave blank to keep current" />
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

        <Card bordered={false} title="Profile Claims" style={{ marginBottom: 24 }}>
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

        <Card bordered={false} title="Contact Information" style={{ marginBottom: 24 }}>
          <Form.Item label="Phone Number" name="phone_number">
            <Input placeholder="+1-234-567-8900" />
          </Form.Item>
        </Card>

        <Card bordered={false} title="Address" style={{ marginBottom: 24 }}>
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
              loading={updateUserMutation.isPending}
            >
              Save Changes
            </Button>
            <Button onClick={() => navigate(`/users/${id}`)}>
              Cancel
            </Button>
          </Space>
        </Card>
      </Form>
    </>
  );
};

export default UserEdit;
