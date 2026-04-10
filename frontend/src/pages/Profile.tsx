import { useState } from 'react';
import { Form, Input, Button, Space, message, Tabs, Spin, Alert } from 'antd';
import { UserOutlined, MailOutlined, LockOutlined, SaveOutlined, IdcardOutlined } from '@ant-design/icons';
import { useProfile, useUpdateProfile, useChangePassword } from '../hooks/useApi';

const SectionCard = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <div style={{ background: 'var(--surface)', borderRadius: 12, border: '1px solid var(--border)', boxShadow: 'var(--shadow-card)', overflow: 'hidden', marginBottom: 24 }}>
    <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)' }}>
      <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-primary)' }}>{title}</span>
    </div>
    <div style={{ padding: '24px' }}>{children}</div>
  </div>
);

const Profile = () => {
  const [profileForm] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [activeTab, setActiveTab] = useState('profile');

  const { data: profile, isLoading, error } = useProfile();
  const updateProfileMutation = useUpdateProfile();
  const changePasswordMutation = useChangePassword();

  const handleProfileSubmit = async (values: { email: string; name: string }) => {
    try {
      await updateProfileMutation.mutateAsync({ email: values.email, name: values.name });
      message.success('Profile updated successfully');
    } catch (err) {
      message.error((err as Error).message || 'Failed to update profile');
    }
  };

  const handlePasswordSubmit = async (values: { currentPassword: string; newPassword: string; confirmPassword: string }) => {
    if (values.newPassword !== values.confirmPassword) {
      message.error('New passwords do not match');
      return;
    }
    try {
      await changePasswordMutation.mutateAsync({ currentPassword: values.currentPassword, newPassword: values.newPassword });
      message.success('Password changed successfully');
      passwordForm.resetFields();
    } catch (err) {
      message.error((err as Error).message || 'Failed to change password');
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
    return <Alert message="Error" description="Failed to load profile" type="error" showIcon />;
  }

  const avatarLetter = (profile.username || 'U')[0].toUpperCase();

  return (
    <>
      <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 28 }}>
        <div style={{ width: 56, height: 56, borderRadius: '50%', background: 'var(--color-primary)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
          <span style={{ fontSize: 22, fontWeight: 700, color: '#fff' }}>{avatarLetter}</span>
        </div>
        <div>
          <div style={{ fontSize: 20, fontWeight: 700, color: 'var(--text-primary)', lineHeight: 1.2 }}>{profile.name || profile.username}</div>
          <div style={{ fontSize: 14, color: 'var(--text-muted)', marginTop: 2 }}>@{profile.username}</div>
        </div>
      </div>

      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        tabBarStyle={{ marginBottom: 24 }}
        items={[
          {
            key: 'profile',
            label: <span><UserOutlined style={{ marginRight: 6 }} />Profile</span>,
            children: (
              <SectionCard title="Profile Information">
                <Form
                  form={profileForm}
                  layout="vertical"
                  initialValues={{ username: profile.username, email: profile.email, name: profile.name, role: profile.role }}
                  onFinish={handleProfileSubmit}
                  style={{ maxWidth: 560 }}
                >
                  <Form.Item label="Username" name="username">
                    <Input prefix={<UserOutlined />} disabled />
                  </Form.Item>
                  <Form.Item label="Role" name="role">
                    <Input prefix={<IdcardOutlined />} disabled />
                  </Form.Item>
                  <Form.Item
                    label="Email"
                    name="email"
                    rules={[
                      { required: true, message: 'Email is required' },
                      { type: 'email', message: 'Please enter a valid email' },
                    ]}
                  >
                    <Input prefix={<MailOutlined />} placeholder="Email address" />
                  </Form.Item>
                  <Form.Item
                    label="Full Name"
                    name="name"
                    rules={[{ required: true, message: 'Name is required' }]}
                  >
                    <Input prefix={<UserOutlined />} placeholder="Full name" />
                  </Form.Item>
                  <Form.Item>
                    <Button
                      type="primary"
                      htmlType="submit"
                      icon={<SaveOutlined />}
                      loading={updateProfileMutation.isPending}
                      style={{ background: 'var(--color-primary)', borderColor: 'var(--color-primary)' }}
                    >
                      Save Changes
                    </Button>
                  </Form.Item>
                </Form>
              </SectionCard>
            ),
          },
          {
            key: 'security',
            label: <span><LockOutlined style={{ marginRight: 6 }} />Security</span>,
            children: (
              <SectionCard title="Change Password">
                <Form
                  form={passwordForm}
                  layout="vertical"
                  onFinish={handlePasswordSubmit}
                  style={{ maxWidth: 560 }}
                >
                  <Form.Item
                    label="Current Password"
                    name="currentPassword"
                    rules={[{ required: true, message: 'Current password is required' }]}
                  >
                    <Input.Password prefix={<LockOutlined />} placeholder="Enter current password" />
                  </Form.Item>
                  <Form.Item
                    label="New Password"
                    name="newPassword"
                    rules={[
                      { required: true, message: 'New password is required' },
                      { min: 6, message: 'Password must be at least 6 characters' },
                    ]}
                  >
                    <Input.Password prefix={<LockOutlined />} placeholder="Enter new password" />
                  </Form.Item>
                  <Form.Item
                    label="Confirm New Password"
                    name="confirmPassword"
                    dependencies={['newPassword']}
                    rules={[
                      { required: true, message: 'Please confirm your new password' },
                      ({ getFieldValue }) => ({
                        validator(_, value) {
                          if (!value || getFieldValue('newPassword') === value) return Promise.resolve();
                          return Promise.reject(new Error('Passwords do not match'));
                        },
                      }),
                    ]}
                  >
                    <Input.Password prefix={<LockOutlined />} placeholder="Confirm new password" />
                  </Form.Item>
                  <Form.Item>
                    <Space>
                      <Button
                        type="primary"
                        htmlType="submit"
                        icon={<LockOutlined />}
                        loading={changePasswordMutation.isPending}
                        style={{ background: 'var(--color-primary)', borderColor: 'var(--color-primary)' }}
                      >
                        Change Password
                      </Button>
                      <Button onClick={() => passwordForm.resetFields()}>Reset</Button>
                    </Space>
                  </Form.Item>
                </Form>
              </SectionCard>
            ),
          },
        ]}
      />
    </>
  );
};

export default Profile;
