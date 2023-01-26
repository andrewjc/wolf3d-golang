import base64
import json

import PIL
import numpy as np
import torch
import torch.nn.functional as ff
import torch.nn as nn
import cv2
from tqdm import tqdm
import joblib

from torch.utils.data import Dataset, DataLoader


useAccelerator = True
IMG_WIDTH = 128
IMG_HEIGHT = 128

def load_text_file(filename, device):
    sampleCount = 0
    samples = []
    with open('../' + filename, 'r') as f:
        for line in f:
            sampleCount += 1
            try:
                samples.append(json.loads(line))
            except Exception as e:
                print(f'Error on line {sampleCount}: {e}')
                pass
    # Create an instance of the dataset
    dataset = ObservationDataset(samples, device)

    print(f'Loaded {len(dataset)} samples from {filename}')

    return dataset


def imgFromStream(msgData):
    # jpeg decompress msgData into numpy array
    img = cv2.imdecode(np.frombuffer(msgData, np.uint8), cv2.IMREAD_GRAYSCALE)

    # resize to 84x84
    img = cv2.resize(img, (IMG_WIDTH, IMG_HEIGHT))

    # normalize to 0-1
    img = img / 255.0

    img = img.reshape(1, IMG_WIDTH, IMG_HEIGHT)

    return img

from torchvision import transforms

# build image transforms to apply to images during training
img_transforms = transforms.Compose([
    transforms.ToPILImage(),
    transforms.RandomAffine(degrees=20),
    transforms.RandomAffine(0, translate=(0.2, 0.2)),
    transforms.RandomAffine(0, shear=20),
    transforms.RandomAffine(0, scale=(0.8, 1.2)),
    transforms.RandomRotation(20),
    transforms.RandomPerspective(distortion_scale=0.2, p=0.5),
    #transforms.ToTensor(),
    transforms.ToTensor(),
    #transforms.Normalize((0.5,), (0.5,))
])


def asimg(src):
    obs1 = base64.b64decode(src)

    obs1 = imgFromStream(obs1)
    #obs1 = img_transforms(obs1)

    #obs1.save('test.png')
    return obs1

class ObservationDataset(Dataset):
    def __init__(self, data, device):
        self.data = data
        self.device = device

    def __len__(self):
        return len(self.data)

    def __getitem__(self, idx):
        observation = self.data[idx]

        # convert textobs and imgobs to tensors
        textobs = torch.tensor(np.array(observation["Obs1"]), dtype=torch.float32, device=device)
        imgobs = torch.tensor(asimg(observation['Obs2']), dtype=torch.float32, device=device)
        actions = torch.tensor(observation["Action"], dtype=torch.long, device=device)
        labels = actions # torch.LongTensor(actions, device=device)

        return textobs, imgobs, labels

# Create data loaders
num_outputs = 7
common_heads = 32
model1 = torch.nn.Sequential(
    torch.nn.Linear(9, 16),
    torch.nn.BatchNorm1d(16),
    torch.nn.ReLU(),
    #torch.nn.Dropout(0.25),

    torch.nn.Linear(16, 32),
    torch.nn.BatchNorm1d(32),
    torch.nn.ReLU(),
    #torch.nn.Dropout(0.25),

    torch.nn.Linear(32, 32),
    torch.nn.BatchNorm1d(32),
    torch.nn.ReLU(),
    #torch.nn.Dropout(0.25),

    torch.nn.Linear(32, common_heads),
    #torch.nn.LayerNorm(common_heads),
)



class ImageObservationHead(nn.Module):
    def __init__(self):
        super(ImageObservationHead, self).__init__()
        # 1st convolutional block
        self.conv1 = nn.Conv2d(in_channels=1, out_channels=32, kernel_size=5, stride=1, padding=1)
        self.batch_norm1 = nn.BatchNorm2d(num_features=32)
        self.relu1 = nn.ReLU()
        self.pool1 = nn.MaxPool2d(kernel_size=3, stride=2)
        self.drop1 = nn.Dropout(0.25)
        self.ln1 = nn.Linear(64, 64)

        # 2nd convolutional block
        self.conv2 = nn.Conv2d(in_channels=32, out_channels=64, kernel_size=3, stride=1, padding=1)
        self.batch_norm2 = nn.BatchNorm2d(num_features=64)
        self.relu2 = nn.ReLU()
        self.pool2 = nn.MaxPool2d(kernel_size=3, stride=2)
        self.drop2 = nn.Dropout(0.25)
        self.ln2 = nn.Linear(64, 64)

        # 3rd convolutional block
        self.conv3 = nn.Conv2d(in_channels=64, out_channels=64, kernel_size=3, stride=1, padding=1)
        self.batch_norm3 = nn.BatchNorm2d(num_features=64)
        self.relu3 = nn.ReLU()
        self.pool3 = nn.MaxPool2d(kernel_size=3, stride=2)
        self.drop3 = nn.Dropout(0.25)
        self.ln3 = nn.Linear(64, 64)


        # Fully connected layer
        self.fc = nn.Linear(10816, common_heads)
        self.fc2 = nn.Linear(256, common_heads)
        self.fcbn1 = nn.BatchNorm1d(num_features=256)
        self.fcbn2 = nn.BatchNorm1d(num_features=32)

        self.drop4 = nn.Dropout(0.25)
        self.relu4 = nn.ReLU()

        self.ln = nn.LayerNorm(common_heads)

        self.avgPool = nn.AvgPool2d(kernel_size=3, stride=2, padding=0)

    def forward(self, x):
        x = self.conv1(x)
        #x = self.batch_norm1(x)
        #x = self.relu1(x)
        x = self.pool1(x)
        #x = self.drop1(x)

        x = self.conv2(x)
        #x = self.batch_norm2(x)
        #x = self.relu2(x)
        x = self.pool2(x)

        x = self.conv3(x)
        #x = self.batch_norm3(x)
        #x = self.relu3(x)
        x = self.pool3(x)

        #x = self.avgPool(x)

        x = torch.flatten(x, 1)

        x = self.fc(x)
        #x = self.fcbn1(x)
        x = self.relu4(x)
        #x = self.drop4(x)
        #x = self.fc2(x)
       # x = self.fcbn2(x)
        #x = self.relu4(x)

        #x = self.ln(x)
        return x


# construct custom model
class CustomModel(torch.nn.Module):
    def __init__(self):
        super().__init__()
        self.model1 = model1
        self.model2 = ImageObservationHead()
        self.fc = torch.nn.Linear(common_heads * 2, 64)
        self.fc2 = torch.nn.Linear(64, num_outputs)
        self.bn = torch.nn.BatchNorm1d(common_heads*2)
        self.bn2 = torch.nn.BatchNorm1d(num_outputs)
        self.relu = torch.nn.ReLU()
        self.drop = torch.nn.Dropout(0.25)

        self.softmax = torch.nn.Softmax(dim=1)

    def forward(self, x1, x2):
        x1 = self.model1(x1)
        x2 = self.model2(x2)
        x = torch.cat([x1, x2], dim=1)

        x = self.fc(x)
        #x = self.bn(x)
        x = self.relu(x)
        #x = self.drop(x)

        x = self.fc2(x)
        #x = self.bn2(x)
        x = self.relu(x)

        x = self.softmax(x)
        return x


# create model
device = torch.device("cpu")
if useAccelerator and torch.backends.mps.is_available():
    print("Using MPS")
    device = torch.device("mps")

model1.to(device)
model = CustomModel()
model.to(device)

batch_size = 256
weight_decay = 1e-4
optimizer = torch.optim.Adam(model.parameters(), lr=1e-3)
loss_fn = torch.nn.CrossEntropyLoss().to(device)

data_loader = DataLoader(load_text_file("train_data.txt", device), batch_size=batch_size, shuffle=True)
data_loader_test = DataLoader(load_text_file("test_data.txt", device), batch_size=batch_size, shuffle=True)


def train_batch(dl, model, loss_fn, optimizer, stats):
    for textobs, imgobs, labels in tqdm(dl):
        # Convert actions to one-hot labels
        # labels = ff.one_hot(actions, num_classes=num_outputs)
        # Clear the gradients
        optimizer.zero_grad()
        # Forward pass
        outputs = model(textobs, imgobs)
        _, preds = torch.max(outputs, 1)
        # Compute the loss
        loss = loss_fn(outputs, labels)
        # Backward pass
        loss.backward()
        # Update the parameters
        optimizer.step()

        # Compute the running loss and accuracy
        stats['running_loss'] += loss.item() * imgobs.size(0)
        stats['running_acc'] += torch.sum(preds == labels.data).item()
        stats['total'] += labels.size(0)
    return stats

def test_batch(dl, model, loss_fn, stats):
    with torch.no_grad():
        model.eval()
        for textobs, imgobs, labels in dl:
            outputs = model(textobs, imgobs)
            _, preds = torch.max(outputs, 1)
            # Compute the loss
            loss = loss_fn(outputs, labels)

            # Compute the running loss and accuracy
            stats['test_running_loss'] += loss.item() * imgobs.size(0)
            stats['test_running_acc'] += torch.sum(preds == labels.data)
            stats['test_total'] += labels.size(0)
    model.train()
    return stats


# Train the network
for epoch in range(1000):
    stats = {'running_loss': 0.0, 'running_acc': 0, 'total': 0}

    stats = train_batch(data_loader, model, loss_fn, optimizer, stats)

    # Compute the average loss and accuracy for the epoch
    epoch_loss = stats['running_loss'] / stats['total']
    epoch_acc = stats['running_acc'] / stats['total']
    print('Epoch: {} Loss: {:.4f} Acc: {:.4f}'.format(epoch, epoch_loss, epoch_acc))

    # save the model
    model = model.to("cpu")
    torch.save(model.state_dict(), 'model.pth')
    joblib.dump(model, 'model.j')
    model.to(device)

    # Test the network
    stats = {'test_running_loss': 0.0, 'test_running_acc': 0, 'test_total': 0}
    stats = test_batch(data_loader_test, model, loss_fn, stats)

    # Compute the average loss and accuracy for the epoch
    epoch_loss = stats['test_running_loss'] / stats['test_total']
    epoch_acc = stats['test_running_acc'] / stats['test_total']
    print('Test: {} Loss: {:.4f} Acc: {:.4f}'.format(epoch, epoch_loss, epoch_acc))
