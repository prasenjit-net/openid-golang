import { useState } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  Space,
  Typography,
  message,
  Divider,
  Spin,
  Alert,
  Tabs,
} from 'antd';
import {
  UserOutlined,
  MailOutlined,
  LockOutlined,
  SaveOutlined,
  IdcardOutlined,
} from '@ant-design/icons';
import { useProfile, useUpdateProfile, useChangePassword } from '../hooks/useApi';

const { Title, Text } = Typography;

const Profile = () => {
  const [profileForm] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [activeTab, setActiveTab] = useState('profile');
  
  const { data: profile, isLoading, error } = useProfile();
  const updateProfileMutation = useUpdateProfile();
  const changePasswordMutation = useChangePassword();

  const handleProfileSubmit = async (values: { email: string; name: string }) => {
    try {
      await updateProfileMutation.mutateAsync({
        email: values.email,
        name: values.name,
      });
      message.success('Profile updated successfully');
    } catch (error) {
      const err = error as Error;
      message.error(err.message || 'Failed to update profile');
    }
  };

  const handlePasswordSubmit = async (values: { currentPassword: string; newPassword: string; confirmPassword: string }) => {
    if (values.newPassword !== values.confirmPassword) {
      message.error('New passwords do not match');
      return;
    }

    try {
      await changePasswordMutation.mutateAsync({
        currentPassword: values.currentPassword,
        newPassword: values.newPassword,
      });
      message.success('Password changed successfully');
      passwordForm.resetFields();
    } catch (error) {
      const err = error as Error;
      message.error(err.message || 'Failed to change password');
    }
  };

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (error || !profile) {
    return (
      <Alert
        message="Error"
        description="Failed to load profile"
        type="error"
        showIcon
      />
    );
  }

  return (
    <>
      <div style={{ marginBottom: 24 }}>
        <Title level={2}>My Profile</Title>
        <Text type="secondary">Manage your account settings and preferences</Text>
      </div>

      <Card bordered={false}>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            {
              key: 'profile',
              label: (
                <span>
                  <UserOutlined />
                  Profile Information
                </span>
              ),
              children: (
                <div style={{ maxWidth: 600 }}>
                  <Form
                    form={profileForm}
                    layout="vertical"
                    initialValues={{
                      username: profile.username,
                      email: profile.email,
                      name: profile.name,
                      role: profile.role,
                    }}
                    onFinish={handleProfileSubmit}
                  >
                    <Form.Item
                      label="Username"
                      name="username"
                    >
                      <Input
                        prefix={<UserOutlined />}
                        disabled
                        placeholder="Username (cannot be changed)"
                      />
                    </Form.Item>

                    <Form.Item
                      label="Role"
                      name="role"
                    >
                      <Input
                        prefix={<IdcardOutlined />}
                        disabled
                        placeholder="Role"
                      />
                    </Form.Item>

                    <Form.Item
                      label="Email"
                      name="email"
                      rules={[
                        { required: true, message: 'Email is required' },
                        { type: 'email', message: 'Please enter a valid email' },
                      ]}
                    >
                      <Input
                        prefix={<MailOutlined />}
                        placeholder="Email address"
                      />
                    </Form.Item>

                    <Form.Item
                      label="Full Name"
                      name="name"
                      rules={[{ required: true, message: 'Name is required' }]}
                    >
                      <Input
                        prefix={<UserOutlined />}
                        placeholder="Full name"
                      />
                    </Form.Item>

                    <Form.Item>
                      <Button
                        type="primary"
                        htmlType="submit"
                        icon={<SaveOutlined />}
                        loading={updateProfileMutation.isPending}
                      >
                        Save Changes
                      </Button>
                    </Form.Item>
                  </Form>
                </div>
              ),
            },
            {
              key: 'security',
              label: (
                <span>
                  <LockOutlined />
                  Security
                </span>
              ),
              children: (
                <div style={{ maxWidth: 600 }}>
                  <Title level={4}>Change Password</Title>
                  <Text type="secondary">
                    Update your password to keep your account secure
                  </Text>
                  
                  <Divider />

                  <Form
                    form={passwordForm}
                    layout="vertical"
                    onFinish={handlePasswordSubmit}
                  >
                    <Form.Item
                      label="Current Password"
                      name="currentPassword"
                      rules={[
                        { required: true, message: 'Current password is required' },
                      ]}
                    >
                      <Input.Password
                        prefix={<LockOutlined />}
                        placeholder="Enter current password"
                      />
                    </Form.Item>

                    <Form.Item
                      label="New Password"
                      name="newPassword"
                      rules={[
                        { required: true, message: 'New password is required' },
                        { min: 6, message: 'Password must be at least 6 characters' },
                      ]}
                    >
                      <Input.Password
                        prefix={<LockOutlined />}
                        placeholder="Enter new password"
                      />
                    </Form.Item>

                    <Form.Item
                      label="Confirm New Password"
                      name="confirmPassword"
                      dependencies={['newPassword']}
                      rules={[
                        { required: true, message: 'Please confirm your new password' },
                        ({ getFieldValue }) => ({
                          validator(_, value) {
                            if (!value || getFieldValue('newPassword') === value) {
                              return Promise.resolve();
                            }
                            return Promise.reject(new Error('Passwords do not match'));
                          },
                        }),
                      ]}
                    >
                      <Input.Password
                        prefix={<LockOutlined />}
                        placeholder="Confirm new password"
                      />
                    </Form.Item>

                    <Form.Item>
                      <Space>
                        <Button
                          type="primary"
                          htmlType="submit"
                          icon={<LockOutlined />}
                          loading={changePasswordMutation.isPending}
                        >
                          Change Password
                        </Button>
                        <Button onClick={() => passwordForm.resetFields()}>
                          Reset
                        </Button>
                      </Space>
                    </Form.Item>
                  </Form>
                </div>
              ),
            },
          ]}
        />
      </Card>
    </>
  );
};

export default Profile;
