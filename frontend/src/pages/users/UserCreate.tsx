import { useNavigate } from 'react-router-dom';
import { Form, Input, Select, Button, Space, message, Row, Col } from 'antd';
import {
  ArrowLeftOutlined,
  UserOutlined,
  IdcardOutlined,
  PhoneOutlined,
  HomeOutlined,
} from '@ant-design/icons';
import { useCreateUser } from '../../hooks/useApi';

const { TextArea } = Input;

const sectionCard = { background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--border)', boxShadow: 'var(--shadow-card)', overflow: 'hidden', marginBottom: 24 } as const;
const sectionHeader = (icon: React.ReactNode, title: string) => (
  <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 8 }}>
    <span style={{ color: 'var(--color-primary)', fontSize: 16, display: 'flex' }}>{icon}</span>
    <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-primary)' }}>{title}</span>
  </div>
);

const UserCreate = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const createUserMutation = useCreateUser();

  const handleSubmit = async (values: { username: string; email: string; password: string; name: string; role: string }) => {
    try {
      const data = await createUserMutation.mutateAsync(values);
      message.success('User created successfully');
      navigate(`/users/${data.id}`);
    } catch (error) {
      message.error('Failed to create user');
      console.error('Failed to create user:', error);
    }
  };

  return (
    <>
      <Button
        type="link"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/users')}
        style={{ padding: 0, marginBottom: 20, color: 'var(--text-secondary)', fontWeight: 500 }}
      >
        Users
      </Button>

      <div style={{ marginBottom: 24 }}>
        <span style={{ fontSize: 20, fontWeight: 700, color: 'var(--text-primary)' }}>Create New User</span>
      </div>

      <Form form={form} layout="vertical" onFinish={handleSubmit} initialValues={{ role: 'user' }}>
        {/* Basic Information */}
        <div style={sectionCard}>
          {sectionHeader(<UserOutlined />, 'Basic Information')}
          <div style={{ padding: '20px 24px' }}>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="Username" name="username" rules={[{ required: true, message: 'Please enter username' }]}>
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
                <Form.Item label="Email" name="email" rules={[{ required: true, message: 'Please enter email' }, { type: 'email', message: 'Please enter a valid email' }]}>
                  <Input />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="Password" name="password" rules={[{ required: true, message: 'Please enter password' }]}>
                  <Input.Password />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item label="Full Name" name="name" rules={[{ required: true, message: 'Please enter name' }]}>
              <Input />
            </Form.Item>
          </div>
        </div>

        {/* Profile Claims */}
        <div style={sectionCard}>
          {sectionHeader(<IdcardOutlined />, 'Profile Claims')}
          <div style={{ padding: '20px 24px' }}>
            <Row gutter={16}>
              <Col span={8}>
                <Form.Item label="Given Name" name="given_name"><Input /></Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label="Family Name" name="family_name"><Input /></Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label="Middle Name" name="middle_name"><Input /></Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="Nickname" name="nickname"><Input /></Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="Preferred Username" name="preferred_username"><Input /></Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="Profile URL" name="profile"><Input placeholder="https://example.com/profile" /></Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="Picture URL" name="picture"><Input placeholder="https://example.com/avatar.jpg" /></Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="Website" name="website"><Input placeholder="https://example.com" /></Form.Item>
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
                <Form.Item label="Birthdate" name="birthdate"><Input placeholder="YYYY-MM-DD" /></Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label="Timezone" name="zoneinfo"><Input placeholder="America/New_York" /></Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label="Locale" name="locale"><Input placeholder="en-US" /></Form.Item>
              </Col>
            </Row>
          </div>
        </div>

        {/* Contact Information */}
        <div style={sectionCard}>
          {sectionHeader(<PhoneOutlined />, 'Contact Information')}
          <div style={{ padding: '20px 24px' }}>
            <Form.Item label="Phone Number" name="phone_number">
              <Input placeholder="+1-234-567-8900" />
            </Form.Item>
          </div>
        </div>

        {/* Address */}
        <div style={sectionCard}>
          {sectionHeader(<HomeOutlined />, 'Address')}
          <div style={{ padding: '20px 24px' }}>
            <Form.Item label="Formatted Address" name={['address', 'formatted']}>
              <TextArea rows={2} placeholder="Full formatted address" />
            </Form.Item>
            <Form.Item label="Street Address" name={['address', 'street_address']}>
              <Input />
            </Form.Item>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="City/Locality" name={['address', 'locality']}><Input /></Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="State/Region" name={['address', 'region']}><Input /></Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="Postal Code" name={['address', 'postal_code']}><Input /></Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="Country" name={['address', 'country']}><Input /></Form.Item>
              </Col>
            </Row>
          </div>
        </div>

        {/* Actions */}
        <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
          <Space>
            <Button onClick={() => navigate('/users')}>Cancel</Button>
            <Button
              type="primary"
              htmlType="submit"
              loading={createUserMutation.isPending}
              style={{ background: 'var(--color-primary)', borderColor: 'var(--color-primary)' }}
            >
              Create User
            </Button>
          </Space>
        </div>
      </Form>
    </>
  );
};

export default UserCreate;
